package rtk

import (
	"sort"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// RTKAgentConfig maps a gentle-ai agent to its rtk init flags.
// Global indicates whether the -g (global hooks) flag is used.
// Windsurf is project-local (no -g).
// Args are the agent-specific flags passed to rtk init (e.g., "--opencode", "--agent cursor").
// Does NOT include "init", "-g", or "--auto-patch" — those are added by InitArgsForAgent.
type RTKAgentConfig struct {
	Global bool
	Args   []string
}

// AgentConfigs returns the agent-to-RTK config mapping.
// Only agents supported by BOTH gentle-ai AND RTK are included.
// Pending agents (kimi, qwen-code, kiro-ide) are absent.
// Cline/Roo Code is supported by RTK but gentle-ai has no AgentID for it yet.
func AgentConfigs() map[model.AgentID]RTKAgentConfig {
	return map[model.AgentID]RTKAgentConfig{
		// Claude Code — RTK default agent (no specific flag needed)
		model.AgentClaudeCode: {Global: true, Args: nil},
		// OpenCode — uses --opencode flag
		model.AgentOpenCode: {Global: true, Args: []string{"--opencode"}},
		// Gemini CLI — uses --gemini flag
		model.AgentGeminiCLI: {Global: true, Args: []string{"--gemini"}},
		// Codex — uses --codex flag
		model.AgentCodex: {Global: true, Args: []string{"--codex"}},
		// Cursor — uses --agent cursor
		model.AgentCursor: {Global: true, Args: []string{"--agent", "cursor"}},
		// VS Code Copilot — RTK default agent (same as claude-code, uses copilot hooks)
		model.AgentVSCodeCopilot: {Global: true, Args: nil},
		// Antigravity — uses --agent antigravity
		model.AgentAntigravity: {Global: true, Args: []string{"--agent", "antigravity"}},
		// Windsurf — project-local hooks (no -g flag)
		model.AgentWindsurf: {Global: false, Args: []string{"--agent", "windsurf"}},
		// Kilocode — uses --agent kilocode
		model.AgentKilocode: {Global: true, Args: []string{"--agent", "kilocode"}},
	}
}

// InitArgsForAgent returns the full argument list for `rtk init` for the given agent.
// Returns nil if the agent is not supported by RTK.
//
// Example outputs:
//
//	claude-code    → ["init", "-g", "--auto-patch"]
//	opencode       → ["init", "-g", "--opencode", "--auto-patch"]
//	cursor         → ["init", "-g", "--agent", "cursor", "--auto-patch"]
//	windsurf       → ["init", "--agent", "windsurf"]  (no -g, no --auto-patch)
func InitArgsForAgent(agentID model.AgentID) []string {
	cfg, ok := AgentConfigs()[agentID]
	if !ok {
		return nil
	}

	args := []string{"init"}
	if cfg.Global {
		args = append(args, "-g")
	}
	args = append(args, cfg.Args...)
	// Windsurf does not support --auto-patch (project-local hooks).
	if cfg.Global {
		args = append(args, "--auto-patch")
	}
	return args
}

// UninstallArgsForAgent returns the full argument list for `rtk init --uninstall`
// for the given agent. Returns nil if the agent is not supported by RTK.
//
// Example outputs:
//
//	claude-code    → ["init", "-g", "--uninstall"]
//	opencode       → ["init", "-g", "--opencode", "--uninstall"]
//	windsurf       → ["init", "--agent", "windsurf", "--uninstall"]  (no -g)
func UninstallArgsForAgent(agentID model.AgentID) []string {
	cfg, ok := AgentConfigs()[agentID]
	if !ok {
		return nil
	}

	args := []string{"init"}
	if cfg.Global {
		args = append(args, "-g")
	}
	args = append(args, cfg.Args...)
	args = append(args, "--uninstall")
	return args
}

// SupportedAgents returns a sorted list of agent IDs supported by RTK.
func SupportedAgents() []model.AgentID {
	configs := AgentConfigs()
	agents := make([]model.AgentID, 0, len(configs))
	for id := range configs {
		agents = append(agents, id)
	}
	sort.Slice(agents, func(i, j int) bool {
		return string(agents[i]) < string(agents[j])
	})
	return agents
}
