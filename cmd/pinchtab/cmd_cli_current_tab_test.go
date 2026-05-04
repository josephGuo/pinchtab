package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestUseLocalTabStateFileDisabledForIdentifiedCallers(t *testing.T) {
	oldAgentID := cliAgentID
	defer func() { cliAgentID = oldAgentID }()

	t.Run("session", func(t *testing.T) {
		cliAgentID = ""
		t.Setenv("PINCHTAB_AGENT_ID", "")
		t.Setenv("PINCHTAB_SESSION", "ses_test")
		if useLocalTabStateFile() {
			t.Fatal("session-authenticated callers should not use local current-tab state")
		}
	})

	t.Run("agent flag", func(t *testing.T) {
		cliAgentID = "agent-flag"
		t.Setenv("PINCHTAB_AGENT_ID", "")
		t.Setenv("PINCHTAB_SESSION", "")
		if useLocalTabStateFile() {
			t.Fatal("--agent-id callers should not use local current-tab state")
		}
	})

	t.Run("agent env", func(t *testing.T) {
		cliAgentID = ""
		t.Setenv("PINCHTAB_AGENT_ID", "agent-env")
		t.Setenv("PINCHTAB_SESSION", "")
		if useLocalTabStateFile() {
			t.Fatal("PINCHTAB_AGENT_ID callers should not use local current-tab state")
		}
	})
}

func TestDefaultTabFlagFromStateOnlyForAnonymousCallers(t *testing.T) {
	oldAgentID := cliAgentID
	defer func() { cliAgentID = oldAgentID }()

	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("PINCHTAB_AGENT_ID", "")
	t.Setenv("PINCHTAB_SESSION", "")
	cliAgentID = ""
	WriteTabStateFile("tab-local")

	anonymous := &cobra.Command{Use: "anonymous"}
	anonymous.Flags().String("tab", "", "Tab ID")
	defaultTabFlagFromState(anonymous)
	if got, _ := anonymous.Flags().GetString("tab"); got != "tab-local" {
		t.Fatalf("anonymous tab flag = %q, want tab-local", got)
	}
	if anonymous.Flags().Changed("tab") {
		t.Fatal("state-backed tab default should not mark --tab as explicitly changed")
	}

	t.Setenv("PINCHTAB_SESSION", "ses_test")
	sessionScoped := &cobra.Command{Use: "session"}
	sessionScoped.Flags().String("tab", "", "Tab ID")
	defaultTabFlagFromState(sessionScoped)
	if got, _ := sessionScoped.Flags().GetString("tab"); got != "" {
		t.Fatalf("session tab flag = %q, want empty", got)
	}

	t.Setenv("PINCHTAB_SESSION", "")
	cliAgentID = "agent-flag"
	agentScoped := &cobra.Command{Use: "agent"}
	agentScoped.Flags().String("tab", "", "Tab ID")
	defaultTabFlagFromState(agentScoped)
	if got, _ := agentScoped.Flags().GetString("tab"); got != "" {
		t.Fatalf("agent tab flag = %q, want empty", got)
	}
}
