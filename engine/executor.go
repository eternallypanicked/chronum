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
	step := node.Step
    fmt.Printf("→ [shell] executing step: %s\n", step.Name)

    var cmd *exec.Cmd
    switch runtime.GOOS {
    case "windows":
        cmd = exec.Command("cmd", "/C", step.Run)
    default:
        cmd = exec.Command("sh", "-c", step.Run)
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

