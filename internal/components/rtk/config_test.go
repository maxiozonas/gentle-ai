package rtk

import (
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestAgentConfigs_AllSupportedAgentsPresent(t *testing.T) {
	configs := AgentConfigs()

	expectedAgents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentGeminiCLI,
		model.AgentCodex,
		model.AgentCursor,
		model.AgentVSCodeCopilot,
		model.AgentAntigravity,
		model.AgentWindsurf,
		model.AgentKilocode,
	}

	for _, agent := range expectedAgents {
		cfg, ok := configs[agent]
		if !ok {
			t.Errorf("AgentConfigs() missing expected agent %q", agent)
			continue
		}
		if !cfg.Global && agent != model.AgentWindsurf {
			t.Errorf("AgentConfigs()[%q].Global = false, want true (only windsurf is project-local)", agent)
		}
	}

	if got := len(configs); got != len(expectedAgents) {
		t.Errorf("AgentConfigs() has %d entries, want %d", got, len(expectedAgents))
	}
}

func TestAgentConfigs_PendingAgentsAbsent(t *testing.T) {
	configs := AgentConfigs()

	pendingAgents := []model.AgentID{
		model.AgentKimi,
		model.AgentQwenCode,
		model.AgentKiroIDE,
	}

	for _, agent := range pendingAgents {
		if _, ok := configs[agent]; ok {
			t.Errorf("AgentConfigs() should NOT contain pending agent %q", agent)
		}
	}
}

func TestAgentConfigs_WindsurfProjectLocal(t *testing.T) {
	configs := AgentConfigs()

	cfg, ok := configs[model.AgentWindsurf]
	if !ok {
		t.Fatal("AgentConfigs() missing windsurf")
	}
	if cfg.Global {
		t.Error("Windsurf should have Global=false (project-local hooks)")
	}
	if !reflect.DeepEqual(cfg.Args, []string{"--agent", "windsurf"}) {
		t.Errorf("Windsurf Args = %v, want [--agent windsurf]", cfg.Args)
	}
}

func TestAgentConfigs_AntigravityIncluded(t *testing.T) {
	configs := AgentConfigs()

	cfg, ok := configs[model.AgentAntigravity]
	if !ok {
		t.Fatal("AgentConfigs() missing antigravity — it IS supported by RTK")
	}
	if !cfg.Global {
		t.Error("Antigravity should have Global=true")
	}
	if !reflect.DeepEqual(cfg.Args, []string{"--agent", "antigravity"}) {
		t.Errorf("Antigravity Args = %v, want [--agent antigravity]", cfg.Args)
	}
}

func TestAgentConfigs_KilocodeIncluded(t *testing.T) {
	configs := AgentConfigs()

	cfg, ok := configs[model.AgentKilocode]
	if !ok {
		t.Fatal("AgentConfigs() missing kilocode — it IS supported by RTK")
	}
	if !cfg.Global {
		t.Error("Kilocode should have Global=true")
	}
	if !reflect.DeepEqual(cfg.Args, []string{"--agent", "kilocode"}) {
		t.Errorf("Kilocode Args = %v, want [--agent kilocode]", cfg.Args)
	}
}

func TestInitArgsForAgent(t *testing.T) {
	tests := []struct {
		name     string
		agentID  model.AgentID
		wantArgs []string
	}{
		{
			name:     "claude-code uses default with auto-patch",
			agentID:  model.AgentClaudeCode,
			wantArgs: []string{"init", "-g", "--auto-patch"},
		},
		{
			name:     "opencode uses --opencode flag",
			agentID:  model.AgentOpenCode,
			wantArgs: []string{"init", "-g", "--opencode", "--auto-patch"},
		},
		{
			name:     "gemini-cli uses --gemini flag",
			agentID:  model.AgentGeminiCLI,
			wantArgs: []string{"init", "-g", "--gemini", "--auto-patch"},
		},
		{
			name:     "codex uses --codex flag",
			agentID:  model.AgentCodex,
			wantArgs: []string{"init", "-g", "--codex", "--auto-patch"},
		},
		{
			name:     "cursor uses --agent cursor",
			agentID:  model.AgentCursor,
			wantArgs: []string{"init", "-g", "--agent", "cursor", "--auto-patch"},
		},
		{
			name:     "vscode-copilot uses default with auto-patch",
			agentID:  model.AgentVSCodeCopilot,
			wantArgs: []string{"init", "-g", "--auto-patch"},
		},
		{
			name:     "antigravity uses --agent antigravity",
			agentID:  model.AgentAntigravity,
			wantArgs: []string{"init", "-g", "--agent", "antigravity", "--auto-patch"},
		},
		{
			name:     "windsurf is project-local (no -g, no --auto-patch)",
			agentID:  model.AgentWindsurf,
			wantArgs: []string{"init", "--agent", "windsurf"},
		},
		{
			name:     "kilocode uses --agent kilocode",
			agentID:  model.AgentKilocode,
			wantArgs: []string{"init", "-g", "--agent", "kilocode", "--auto-patch"},
		},
		{
			name:     "unsupported agent returns nil",
			agentID:  model.AgentKimi,
			wantArgs: nil,
		},
		{
			name:     "pending agent qwen-code returns nil",
			agentID:  model.AgentQwenCode,
			wantArgs: nil,
		},
		{
			name:     "pending agent kiro-ide returns nil",
			agentID:  model.AgentKiroIDE,
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InitArgsForAgent(tt.agentID)
			if !reflect.DeepEqual(got, tt.wantArgs) {
				t.Errorf("InitArgsForAgent(%q) = %v, want %v", tt.agentID, got, tt.wantArgs)
			}
		})
	}
}

func TestUninstallArgsForAgent(t *testing.T) {
	tests := []struct {
		name     string
		agentID  model.AgentID
		wantArgs []string
	}{
		{
			name:     "claude-code uninstall",
			agentID:  model.AgentClaudeCode,
			wantArgs: []string{"init", "-g", "--uninstall"},
		},
		{
			name:     "opencode uninstall",
			agentID:  model.AgentOpenCode,
			wantArgs: []string{"init", "-g", "--opencode", "--uninstall"},
		},
		{
			name:     "gemini-cli uninstall",
			agentID:  model.AgentGeminiCLI,
			wantArgs: []string{"init", "-g", "--gemini", "--uninstall"},
		},
		{
			name:     "codex uninstall",
			agentID:  model.AgentCodex,
			wantArgs: []string{"init", "-g", "--codex", "--uninstall"},
		},
		{
			name:     "cursor uninstall",
			agentID:  model.AgentCursor,
			wantArgs: []string{"init", "-g", "--agent", "cursor", "--uninstall"},
		},
		{
			name:     "vscode-copilot uninstall",
			agentID:  model.AgentVSCodeCopilot,
			wantArgs: []string{"init", "-g", "--uninstall"},
		},
		{
			name:     "antigravity uninstall",
			agentID:  model.AgentAntigravity,
			wantArgs: []string{"init", "-g", "--agent", "antigravity", "--uninstall"},
		},
		{
			name:     "windsurf uninstall (no -g)",
			agentID:  model.AgentWindsurf,
			wantArgs: []string{"init", "--agent", "windsurf", "--uninstall"},
		},
		{
			name:     "kilocode uninstall",
			agentID:  model.AgentKilocode,
			wantArgs: []string{"init", "-g", "--agent", "kilocode", "--uninstall"},
		},
		{
			name:     "unsupported agent returns nil",
			agentID:  model.AgentKimi,
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UninstallArgsForAgent(tt.agentID)
			if !reflect.DeepEqual(got, tt.wantArgs) {
				t.Errorf("UninstallArgsForAgent(%q) = %v, want %v", tt.agentID, got, tt.wantArgs)
			}
		})
	}
}

func TestSupportedAgents(t *testing.T) {
	agents := SupportedAgents()

	if len(agents) != 9 {
		t.Fatalf("SupportedAgents() returned %d agents, want 9", len(agents))
	}

	// Verify sorted order
	for i := 1; i < len(agents); i++ {
		if string(agents[i-1]) >= string(agents[i]) {
			t.Errorf("SupportedAgents() not sorted: %q >= %q", agents[i-1], agents[i])
		}
	}
}

func TestInitArgs_ClaudeAndCopilotBothDefault(t *testing.T) {
	// Claude Code and VS Code Copilot both use RTK's default mode (no agent-specific flag).
	// They should produce the same InitArgs.
	claudeArgs := InitArgsForAgent(model.AgentClaudeCode)
	copilotArgs := InitArgsForAgent(model.AgentVSCodeCopilot)

	if !reflect.DeepEqual(claudeArgs, copilotArgs) {
		t.Errorf("claude-code args %v != vscode-copilot args %v — both should use RTK default",
			claudeArgs, copilotArgs)
	}
}
