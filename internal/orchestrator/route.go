package orchestrator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pinchtab/pinchtab/internal/activity"
	"github.com/pinchtab/pinchtab/internal/bridge"
	"github.com/pinchtab/pinchtab/internal/handlers"
	"github.com/pinchtab/pinchtab/internal/httpx"
	"github.com/pinchtab/pinchtab/internal/session"
)

// RoutingDecision identifies which precedence rule selected the target
// instance for a shorthand request. Used for metrics and logging.
type RoutingDecision string

const (
	RoutingDecisionTabOwner RoutingDecision = "tab_owner"
	RoutingDecisionSession  RoutingDecision = "session"
	RoutingDecisionAgent    RoutingDecision = "agent"
	RoutingDecisionFallback RoutingDecision = "fallback"
)

// WrapShorthand wraps a strategy-supplied fallback handler with the routing
// precedence defined in tab-spec.md:
//
//  1. Explicit tab id (path / query / JSON body): route to the owning
//     instance. Not found + multiple instances → 404; not found + single
//     instance → fall through to the only instance for legacy ergonomics.
//  2. Session binding: route to the bound instance if still running.
//  3. Agent binding: route to the bound instance if still running.
//  4. Fallback: invoke the strategy handler.
//
// The wrapper does not write or move bindings — that is step 1.5's job and
// happens in the proxy response hook after a successful response.
func (o *Orchestrator) WrapShorthand(fallback http.HandlerFunc) http.HandlerFunc {
	if o == nil {
		return fallback
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if tabID, _ := ExtractExplicitTabID(r); tabID != "" {
			if o.routeByTabOwner(w, r, tabID) {
				return
			}
			// routeByTabOwner already wrote a 404 / single-instance proxy;
			// only return false when no decision could be made.
			return
		}

		if instanceID, decision, ok := o.resolveIdentityBinding(r); ok {
			if o.routeToInstanceID(w, r, instanceID, decision) {
				return
			}
		}

		fallback(w, r)
	}
}

// routeByTabOwner handles precedence rule 1. Returns true if a response was
// already written (either a successful proxy or an error).
func (o *Orchestrator) routeByTabOwner(w http.ResponseWriter, r *http.Request, tabID string) bool {
	// Fast path via locator cache.
	if o.instanceMgr != nil {
		if inst, err := o.instanceMgr.FindInstanceByTabID(tabID); err == nil && inst != nil {
			if !o.allowCrossInstance(w, r, inst.ID) {
				return true
			}
			o.proxyToInstanceForRoute(w, r, inst, tabID, RoutingDecisionTabOwner)
			return true
		}
	}
	// Slow path: enumerate running instances.
	if internal, err := o.findRunningInstanceByTabID(tabID); err == nil && internal != nil {
		if o.instanceMgr != nil {
			o.instanceMgr.Locator.Register(tabID, internal.ID)
		}
		if !o.allowCrossInstance(w, r, internal.ID) {
			return true
		}
		o.proxyToInstanceForRoute(w, r, &internal.Instance, tabID, RoutingDecisionTabOwner)
		return true
	}

	// Tab not found. If exactly one instance is running, fall through to it
	// — preserves legacy ergonomics for users running `--tab` against a
	// just-created tab whose id has not propagated to the dashboard yet.
	if only := o.singleRunningInstance(); only != nil {
		o.proxyToInstanceForRoute(w, r, &only.Instance, tabID, RoutingDecisionFallback)
		return true
	}

	// Multiple instances and no owner found — refuse to guess.
	httpx.Error(w, http.StatusNotFound, fmt.Errorf("tab %q not found", tabID))
	return true
}

// resolveIdentityBinding implements precedence rules 2 and 3. Public
// session-authenticated requests resolve from the authenticated session
// context. Trusted internal proxy hops resolve from the propagated session
// header. Raw public X-PinchTab-Session-Id is never trusted.
func (o *Orchestrator) resolveIdentityBinding(r *http.Request) (string, RoutingDecision, bool) {
	if r == nil || o.bindings == nil {
		return "", "", false
	}
	if id := sessionIDForRouting(r); id != "" {
		if inst, ok := o.bindings.ResolveSession(id); ok {
			return inst, RoutingDecisionSession, true
		}
	}
	if id := strings.TrimSpace(r.Header.Get(activity.HeaderAgentID)); id != "" {
		if inst, ok := o.bindings.ResolveAgent(id); ok {
			return inst, RoutingDecisionAgent, true
		}
	}
	return "", "", false
}

func sessionIDForRouting(r *http.Request) string {
	if r == nil {
		return ""
	}
	if sess, ok := session.FromRequest(r); ok && sess != nil {
		if id := strings.TrimSpace(sess.ID); id != "" {
			return id
		}
	}
	if handlers.IsTrustedInternalProxy(r) {
		return strings.TrimSpace(r.Header.Get(activity.HeaderPTSessionID))
	}
	return ""
}

// routeToInstanceID validates the bound instance is still running, then
// proxies. Stale bindings (instance gone) are cleared and a fall-through
// signal is returned (false means "let fallback handle it").
func (o *Orchestrator) routeToInstanceID(w http.ResponseWriter, r *http.Request, instanceID string, decision RoutingDecision) bool {
	o.mu.RLock()
	internal, ok := o.instances[instanceID]
	o.mu.RUnlock()
	if !ok || internal == nil || internal.Status != "running" || !instanceIsActive(internal) {
		// Stale binding: clear and let fallback decide.
		if o.bindings != nil {
			o.bindings.ClearInstance(instanceID)
		}
		return false
	}
	o.proxyToInstanceForRoute(w, r, &internal.Instance, "", decision)
	return true
}

// allowCrossInstance decides whether a tab-owner-routed request is allowed
// to land on ownerID when the caller's identity is currently bound to a
// different instance. Returns false (and writes a 409) only when strict
// cross-instance routing is enabled. The default rule rebinds silently;
// the actual rebind happens in the proxy response hook on success.
func (o *Orchestrator) allowCrossInstance(w http.ResponseWriter, r *http.Request, ownerID string) bool {
	o.mu.RLock()
	strict := o.strictCrossInstanceTab
	o.mu.RUnlock()
	if !strict || o.bindings == nil || ownerID == "" {
		return true
	}
	if id := sessionIDForRouting(r); id != "" {
		if existing, ok := o.bindings.ResolveSession(id); ok && existing != ownerID {
			httpx.ErrorCode(w, http.StatusConflict, "cross_instance_tab",
				fmt.Sprintf("session %q is bound to instance %q; tab lives on %q", id, existing, ownerID),
				false, nil)
			return false
		}
	}
	if id := strings.TrimSpace(r.Header.Get(activity.HeaderAgentID)); id != "" {
		if existing, ok := o.bindings.ResolveAgent(id); ok && existing != ownerID {
			httpx.ErrorCode(w, http.StatusConflict, "cross_instance_tab",
				fmt.Sprintf("agent %q is bound to instance %q; tab lives on %q", id, existing, ownerID),
				false, nil)
			return false
		}
	}
	return true
}

// proxyToInstanceForRoute handles activity enrichment and the actual proxy
// for a shorthand request that the routing layer resolved to a specific
// instance. tabID may be empty when routing came from an identity binding.
func (o *Orchestrator) proxyToInstanceForRoute(w http.ResponseWriter, r *http.Request, inst *bridge.Instance, tabID string, decision RoutingDecision) {
	if inst == nil {
		httpx.Error(w, http.StatusInternalServerError, fmt.Errorf("nil instance for routing decision %q", decision))
		return
	}
	activity.EnrichRouteActivity(r)
	update := activity.Update{
		InstanceID:  inst.ID,
		ProfileID:   inst.ProfileID,
		ProfileName: inst.ProfileName,
	}
	if tabID != "" {
		update.TabID = tabID
	}
	activity.EnrichRequest(r, update)

	targetURL, err := o.instancePathURLFromBridge(inst, r.URL.Path, r.URL.RawQuery)
	if err != nil {
		httpx.Error(w, http.StatusBadGateway, err)
		return
	}
	o.proxyToURL(w, r, targetURL)
}
