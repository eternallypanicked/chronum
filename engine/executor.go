package engine

import (
	"chronum/parser/dag"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Executor defines how a step is executed.
type Executor interface {
	Execute(node *dag.Node) error
}

// ShellExecutor is a basic implementation using system shell commands.
type ShellExecutor struct  { Timeout time.Duration }
type PythonExecutor struct { Timeout time.Duration }

func (s *ShellExecutor) Execute(node *dag.Node) error {
	step := node.Step
	if step == nil {
		return fmt.Errorf("no step for node %s", node.ID)
	}
	runCmd := strings.TrimSpace(step.Run)
	if runCmd == "" {
		return fmt.Errorf("empty run command for step %s", step.Name)
	}

	// minimal sanitation: avoid control characters that are suspicious;
	// this is NOT a security boundary for untrusted input.
	if strings.ContainsAny(runCmd, "\x00") {
		return fmt.Errorf("invalid characters in run command")
	}

	// prepare context with timeout if configured
	ctx := context.Background()
	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout)
		defer cancel()
	}

	fmt.Printf("→ [shell] executing step: %s\n", step.Name)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "cmd", "/C", runCmd)
	default:
		// Use /bin/sh -c for POSIX. If you accept untrusted input, run inside a sandbox/container instead.
		cmd = exec.CommandContext(ctx, "sh", "-c", runCmd)
	}

	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", string(output))
	if err != nil {
		// If context deadline exceeded, wrap for clarity
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("step timeout: %w", err)
		}
		return err
	}
	return nil
}

func (p *PythonExecutor) Execute(node *dag.Node) error {
	// For now we expect node.Step.Run to contain the python script path or command.
	step := node.Step
	if step == nil {
		return fmt.Errorf("no step for node %s", node.ID)
	}
	runCmd := strings.TrimSpace(step.Run)
	if runCmd == "" {
		// fallback: if no Run defined, try node.ID (older code used node.ID)
		runCmd = node.ID
	}

	// Use exec.CommandContext to allow cancellation and timeout
	ctx := context.Background()
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}

	fmt.Printf("→ [python] executing step: %s\n", node.ID)
	cmd := exec.CommandContext(ctx, "python3", "-u", "-c", runCmd)
	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", string(output))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("python step timeout: %w", err)
		}
		return err
	}
	return nil
}

