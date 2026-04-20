package rtk

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/internal/model"
)

func TestInject_SingleAgent(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var capturedArgs []string
	runRTKInit = func(args ...string) error {
		capturedArgs = args
		return nil
	}

	result, err := Inject([]model.AgentID{model.AgentOpenCode})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	wantConfigured := []model.AgentID{model.AgentOpenCode}
	if !reflect.DeepEqual(result.Configured, wantConfigured) {
		t.Errorf("Inject().Configured = %v, want %v", result.Configured, wantConfigured)
	}

	wantArgs := []string{"init", "-g", "--opencode", "--auto-patch"}
	if !reflect.DeepEqual(capturedArgs, wantArgs) {
		t.Errorf("rtk init args = %v, want %v", capturedArgs, wantArgs)
	}
}

func TestInject_MultipleAgents(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callLog [][]string
	runRTKInit = func(args ...string) error {
		callLog = append(callLog, args)
		return nil
	}

	agents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentCursor,
	}

	result, err := Inject(agents)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	if !reflect.DeepEqual(result.Configured, agents) {
		t.Errorf("Inject().Configured = %v, want %v", result.Configured, agents)
	}

	if len(callLog) != 3 {
		t.Fatalf("expected 3 rtk init calls, got %d", len(callLog))
	}

	// Verify each call had correct args
	wantCalls := [][]string{
		{"init", "-g", "--auto-patch"},
		{"init", "-g", "--opencode", "--auto-patch"},
		{"init", "-g", "--agent", "cursor", "--auto-patch"},
	}
	for i, want := range wantCalls {
		if !reflect.DeepEqual(callLog[i], want) {
			t.Errorf("call %d: args = %v, want %v", i, callLog[i], want)
		}
	}
}

func TestInject_SkipsUnsupportedAgents(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callCount int
	runRTKInit = func(args ...string) error {
		callCount++
		return nil
	}

	agents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentKimi,     // pending — skip
		model.AgentQwenCode, // pending — skip
		model.AgentKiroIDE,  // pending — skip
		model.AgentOpenCode,
	}

	result, err := Inject(agents)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	wantConfigured := []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}
	if !reflect.DeepEqual(result.Configured, wantConfigured) {
		t.Errorf("Inject().Configured = %v, want %v", result.Configured, wantConfigured)
	}

	if callCount != 2 {
		t.Errorf("rtk init called %d times, want 2 (skipped 3 pending agents)", callCount)
	}
}

func TestInject_WindsurfNoGlobalFlag(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var capturedArgs []string
	runRTKInit = func(args ...string) error {
		capturedArgs = args
		return nil
	}

	result, err := Inject([]model.AgentID{model.AgentWindsurf})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	if !reflect.DeepEqual(result.Configured, []model.AgentID{model.AgentWindsurf}) {
		t.Errorf("Inject().Configured = %v, want [windsurf]", result.Configured)
	}

	// Windsurf: no -g, no --auto-patch
	wantArgs := []string{"init", "--agent", "windsurf"}
	if !reflect.DeepEqual(capturedArgs, wantArgs) {
		t.Errorf("windsurf args = %v, want %v", capturedArgs, wantArgs)
	}
}

func TestInject_PartialFailure(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	callCount := 0
	runRTKInit = func(args ...string) error {
		callCount++
		if callCount == 2 {
			return fmt.Errorf("simulated failure")
		}
		return nil
	}

	agents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentCursor,
	}

	result, err := Inject(agents)
	if err == nil {
		t.Fatal("Inject() should return error on partial failure")
	}

	// First agent should be configured, second failed, third not attempted
	wantConfigured := []model.AgentID{model.AgentClaudeCode}
	if !reflect.DeepEqual(result.Configured, wantConfigured) {
		t.Errorf("Inject().Configured = %v, want %v (partial success before failure)", result.Configured, wantConfigured)
	}

	// Should have stopped after failure — only 2 calls (1 success + 1 failure)
	if callCount != 2 {
		t.Errorf("rtk init called %d times, want 2", callCount)
	}
}

func TestInject_EmptyInput(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callCount int
	runRTKInit = func(args ...string) error {
		callCount++
		return nil
	}

	result, err := Inject([]model.AgentID{})
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if len(result.Configured) != 0 {
		t.Errorf("Inject().Configured = %v, want empty", result.Configured)
	}
	if callCount != 0 {
		t.Errorf("rtk init called %d times on empty input, want 0", callCount)
	}
}

func TestInject_AllUnsupported(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callCount int
	runRTKInit = func(args ...string) error {
		callCount++
		return nil
	}

	agents := []model.AgentID{
		model.AgentKimi,
		model.AgentQwenCode,
		model.AgentKiroIDE,
	}

	result, err := Inject(agents)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}
	if len(result.Configured) != 0 {
		t.Errorf("Inject().Configured = %v, want empty (all unsupported)", result.Configured)
	}
	if callCount != 0 {
		t.Errorf("rtk init called %d times for all-unsupported input, want 0", callCount)
	}
}

func TestInjectForUninstall_Success(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callLog [][]string
	runRTKInit = func(args ...string) error {
		callLog = append(callLog, args)
		return nil
	}

	agents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
	}

	result, errs := InjectForUninstall(agents)
	if len(errs) != 0 {
		t.Fatalf("InjectForUninstall() errors = %v", errs)
	}

	wantConfigured := []model.AgentID{model.AgentClaudeCode, model.AgentOpenCode}
	if !reflect.DeepEqual(result.Configured, wantConfigured) {
		t.Errorf("InjectForUninstall().Configured = %v, want %v", result.Configured, wantConfigured)
	}

	wantCalls := [][]string{
		{"init", "-g", "--uninstall"},
		{"init", "-g", "--opencode", "--uninstall"},
	}
	for i, want := range wantCalls {
		if !reflect.DeepEqual(callLog[i], want) {
			t.Errorf("call %d: args = %v, want %v", i, callLog[i], want)
		}
	}
}

func TestInjectForUninstall_PartialFailure(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	callCount := 0
	runRTKInit = func(args ...string) error {
		callCount++
		if callCount == 1 {
			return fmt.Errorf("uninstall failed for agent 1")
		}
		return nil
	}

	agents := []model.AgentID{
		model.AgentClaudeCode,
		model.AgentOpenCode,
		model.AgentCursor,
	}

	result, errs := InjectForUninstall(agents)

	// Uninstall continues through all agents, collecting errors
	if len(errs) != 1 {
		t.Fatalf("InjectForUninstall() errors = %v, want 1 error", errs)
	}

	// Agent 1 failed, but agents 2 and 3 should succeed
	wantConfigured := []model.AgentID{model.AgentOpenCode, model.AgentCursor}
	if !reflect.DeepEqual(result.Configured, wantConfigured) {
		t.Errorf("InjectForUninstall().Configured = %v, want %v", result.Configured, wantConfigured)
	}

	if callCount != 3 {
		t.Errorf("rtk uninstall called %d times, want 3 (continues after failure)", callCount)
	}
}

func TestInjectForUninstall_WindsurfNoGlobal(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var capturedArgs []string
	runRTKInit = func(args ...string) error {
		capturedArgs = args
		return nil
	}

	_, errs := InjectForUninstall([]model.AgentID{model.AgentWindsurf})
	if len(errs) != 0 {
		t.Fatalf("InjectForUninstall() errors = %v", errs)
	}

	wantArgs := []string{"init", "--agent", "windsurf", "--uninstall"}
	if !reflect.DeepEqual(capturedArgs, wantArgs) {
		t.Errorf("windsurf uninstall args = %v, want %v", capturedArgs, wantArgs)
	}
}

func TestInjectForUninstall_SkipsUnsupported(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callCount int
	runRTKInit = func(args ...string) error {
		callCount++
		return nil
	}

	agents := []model.AgentID{
		model.AgentKimi,
		model.AgentClaudeCode,
		model.AgentQwenCode,
	}

	result, errs := InjectForUninstall(agents)
	if len(errs) != 0 {
		t.Fatalf("InjectForUninstall() errors = %v", errs)
	}

	if !reflect.DeepEqual(result.Configured, []model.AgentID{model.AgentClaudeCode}) {
		t.Errorf("InjectForUninstall().Configured = %v, want [claude-code]", result.Configured)
	}

	if callCount != 1 {
		t.Errorf("rtk uninstall called %d times, want 1 (skipped 2 pending agents)", callCount)
	}
}

func TestInject_AllNineAgents(t *testing.T) {
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	var callLog [][]string
	runRTKInit = func(args ...string) error {
		callLog = append(callLog, args)
		return nil
	}

	agents := SupportedAgents()
	result, err := Inject(agents)
	if err != nil {
		t.Fatalf("Inject() error = %v", err)
	}

	if len(result.Configured) != len(agents) {
		t.Errorf("Inject() configured %d agents, want %d", len(result.Configured), len(agents))
	}

	if len(callLog) != len(agents) {
		t.Errorf("rtk init called %d times, want %d", len(callLog), len(agents))
	}
}

func TestInject_CorrectArgsPerAgent(t *testing.T) {
	// Exhaustive test: verify every supported agent produces the correct args.
	origRunRTKInit := runRTKInit
	t.Cleanup(func() { runRTKInit = origRunRTKInit })

	captured := map[model.AgentID][]string{}
	runRTKInit = func(args ...string) error {
		// We can't easily determine which agent from args alone,
		// so let's test each agent individually.
		return nil
	}

	// Test each agent individually
	for _, agentID := range SupportedAgents() {
		agent := agentID // capture
		var got []string
		runRTKInit = func(args ...string) error {
			got = args
			return nil
		}

		_, err := Inject([]model.AgentID{agent})
		if err != nil {
			t.Errorf("Inject(%q) error = %v", agent, err)
			continue
		}

		want := InitArgsForAgent(agent)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Inject(%q) args = %v, want %v", agent, got, want)
		}

		captured[agent] = got
	}

	// Make sure we didn't miss any
	if len(captured) != 9 {
		t.Errorf("captured args for %d agents, want 9", len(captured))
	}
}
