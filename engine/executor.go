package engine

import (
	"chronum/parser/dag"
	"fmt"
	"os/exec"
	"runtime"
)

// Executor defines how a step is executed.
type Executor interface {
	Execute(node *dag.Node) error
}

// ShellExecutor is a basic implementation using system shell commands.
type ShellExecutor struct{}
type PythonExecutor struct{}

func (s *ShellExecutor) Execute(node *dag.Node) error {
	fmt.Printf("→ [shell] executing step: %s\n", node.ID)

	// TODO: In future, node should hold a reference to its StepDefinition
	command := fmt.Sprintf("echo Running %s", node.ID)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", command)
	default:
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", string(output))
	return err
}

func (p *PythonExecutor) Execute(node *dag.Node) error {
    fmt.Printf("→ [python] executing step: %s\n", node.ID)
    cmd := exec.Command("python3", node.ID) // node.ID placeholder, replace with step.Run soon
    output, err := cmd.CombinedOutput()
    fmt.Printf("%s\n", string(output))
    return err
}

