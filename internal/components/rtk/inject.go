package rtk

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

// InjectResult describes the outcome of configuring RTK for a set of agents.
type InjectResult struct {
	// Configured holds the agent IDs that were successfully configured.
	Configured []model.AgentID
	// Skipped holds the agent IDs that were requested but not configured because
	// upstream rtk does not support them yet (e.g. Kimi, Qwen Code, Kiro IDE).
	// The install pipeline surfaces these in the final report so users know
	// which selected agents missed out on token savings and why.
	Skipped []model.AgentID
}

// runRTKInit is the function used to execute `rtk init` commands.
// Package-level var for testability — tests can replace this to avoid real subprocess calls.
var runRTKInit = defaultRunRTKInit

// defaultRunRTKInit executes `rtk` with the given arguments and a 30-second timeout.
func defaultRunRTKInit(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return exec.CommandContext(ctx, "rtk", args...).Run()
}

// Inject configures RTK for the given agents by calling `rtk init` for each one.
// Unsupported agents are silently skipped.
// On failure for any agent, the function returns the partial result (successfully
// configured agents) along with the error for the failing agent.
func Inject(agentIDs []model.AgentID) (InjectResult, error) {
	configs := AgentConfigs()
	var configured, skipped []model.AgentID

	for _, id := range agentIDs {
		if _, ok := configs[id]; !ok {
			// Agent not supported by RTK upstream — record it so the run
			// pipeline can surface a single "skipped N agents" message.
			skipped = append(skipped, id)
			continue
		}

		args := InitArgsForAgent(id)
		if args == nil {
			// Should not happen if config exists, but guard anyway.
			continue
		}

		if err := runRTKInit(args...); err != nil {
			return InjectResult{Configured: configured, Skipped: skipped},
				fmt.Errorf("rtk init for %q: %w", id, err)
		}

		configured = append(configured, id)
	}

	return InjectResult{Configured: configured, Skipped: skipped}, nil
}

// InjectForUninstall runs `rtk init --uninstall` for each given agent.
// Unlike Inject, this continues through all agents even if one fails,
// collecting errors rather than stopping at the first failure.
// Returns the agents that were successfully uninstalled and any errors encountered.
func InjectForUninstall(agentIDs []model.AgentID) (InjectResult, []error) {
	configs := AgentConfigs()
	var configured []model.AgentID
	var errs []error

	for _, id := range agentIDs {
		if _, ok := configs[id]; !ok {
			continue
		}

		args := UninstallArgsForAgent(id)
		if args == nil {
			continue
		}

		if err := runRTKInit(args...); err != nil {
			errs = append(errs, fmt.Errorf("rtk uninstall for %q: %w", id, err))
			continue
		}

		configured = append(configured, id)
	}

	return InjectResult{Configured: configured}, errs
}
