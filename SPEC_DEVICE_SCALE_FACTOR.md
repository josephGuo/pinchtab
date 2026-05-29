# Spec — tab-level `deviceScaleFactor` emulation

Status: draft. Not implemented. Tracking for a separate PR after
`feat/add-capture-endpoint` lands.

## Problem

PinchTab today runs against the host's real Chrome + real display, so
`window.devicePixelRatio` is whatever the host gives you (1 on most Linux
servers, 2 on retina Macs, 1.5 on common Windows hardware, …). Screenshots
come back at that native DPR. Two consequences for agent workflows:

1. **Non-portable image dimensions.** The same `pinchtab capture` against
   the same fixture returns different pixel sizes on different hosts —
   the LLM has to figure out scale from `image.devicePixelRatio` every
   time. Most agent frameworks pin DPR so this just stops mattering.
2. **Bounding-box ↔ image-pixel translation lives in the client.** We
   already return `image.devicePixelRatio` so a vision agent can rescale
   if it wants, but the math is busywork the daemon could do upstream.

The standard fix used by browser-use, Playwright, and Puppeteer is to set
`deviceScaleFactor` once at browser-context creation. Every subsequent
screenshot inherits it. No per-call knob.

## Proposed contract

A new tab-level setting `deviceScaleFactor: float` controls the DPR used
for all rendering and capture within that tab. Default = native (no
emulation).

When set to N > 0:

- `Emulation.setDeviceMetricsOverride({width, height, deviceScaleFactor: N, mobile: false})`
  is called at tab attach, before the first navigation, against the tab
  context.
- `window.devicePixelRatio` returns N inside the page.
- `/screenshot`, `/capture`, screencast frames are all rasterized at N CSS
  pixels → N image pixels per axis.
- Bounding boxes from `DOM.getBoxModel` stay in CSS pixels (DPR doesn't
  change CSS coordinates).
- Click, type, navigation are unchanged (CSS-pixel based).

`Emulation.clearDeviceMetricsOverride` reverts.

## What it does NOT change

- Bounding boxes (`A11yNode.BoundingBox`): always CSS pixels.
- Click / type / scroll coordinates: always CSS pixels.
- Accessibility tree: unaffected.
- The `image.devicePixelRatio` field in `/capture` response: stays
  accurate — reflects the emulated DPR when override is active, native
  DPR otherwise.

## API surface

### Tab creation

```bash
pinchtab tab new --device-scale-factor 1
```

```http
POST /tab
{
  "action": "new",
  "url": "...",
  "deviceScaleFactor": 1
}
```

### Runtime control on an existing tab

```bash
pinchtab tab emulate --tab-id <x> --device-scale-factor 2
pinchtab tab emulate --tab-id <x> --clear
```

```http
POST /tabs/{id}/emulate
{ "deviceScaleFactor": 2, "width": 1440, "height": 900 }

DELETE /tabs/{id}/emulate
```

`width` and `height` are required by CDP for the override call. If
omitted, the implementation reads the current viewport and reuses it.

### Profile / instance default

Set in profile config so all tabs spun up for an instance inherit:

```yaml
deviceScaleFactor: 1
```

Tabs adopted from a real user profile keep the user's native DPR unless
the instance config overrides.

## Stealth posture

`window.devicePixelRatio` is fingerprintable. Emulating DPR=1 on retina
hardware is a divergence sites can detect. The stealth implications:

- Default = no emulation (preserve real-profile property).
- Agent-instance profiles set explicitly when predictability matters
  more than stealth.
- Tabs marked for human handoff should never have an override active —
  the human sees a re-rasterized page.

If the override is set mid-session it triggers a reflow, fires
`IntersectionObserver` callbacks, and may resolve `loading="lazy"`
images. Same hazard class as `--beyond-viewport`. Recommended pattern:
set at tab init only; runtime control reserved for explicit operator
action.

## Out of scope (separate work)

- Per-call DPR override on `/screenshot` / `/capture`. Use `?scale=` for
  bitmap rescale (already shipped); use tab-level emulation when you want
  the page to *render* at a different DPR.
- Touch / mobile emulation. `Emulation.setDeviceMetricsOverride` accepts
  a `mobile` flag and viewport-orientation params — those belong in a
  separate "mobile emulation" spec.
- Network throttling, locale, timezone. Same emulation domain but
  different concerns.

## Implementation footprint

- `bridge.TabPolicyState` or profile-config struct: add
  `DeviceScaleFactor float64` (zero = no override).
- `bridge/tab_manager.go` tab-attach path: when set, call
  `Emulation.setDeviceMetricsOverride` before first navigation.
- New handler `POST /tabs/{id}/emulate` and `DELETE
  /tabs/{id}/emulate`.
- CLI: `pinchtab tab new --device-scale-factor`, `pinchtab tab emulate`.
- MCP tool: extend `pinchtab_list_tabs` response with the active
  emulation, optional `pinchtab_emulate` tool for setting it.

Estimated ~150 LOC plus tests. Independent of any other in-flight work.

## Open questions

- Should we surface DPR as a column in `pinchtab tabs` output? Probably
  yes — agents need to know what they're capturing into.
- Default for `pinchtab tab new` when invoked without flags: native DPR
  or 1? Native DPR preserves backwards compat. 1 makes agent workflows
  predictable. Lean native; agent operators set 1 explicitly.
- Interaction with `--beyond-viewport`: should we re-call
  `setDeviceMetricsOverride` to grow the layout viewport, or let
  `captureBeyondViewport` handle it as today? Today's mechanism is
  already a layout expansion — they're orthogonal but combining them
  needs a test.
