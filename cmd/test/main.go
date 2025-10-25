package main

import (
    "fmt"
    "os"

    "chronum/parser"
    "chronum/parser/types"
    _ "chronum/parser/yaml" // side-effect import for YAML parser

    "chronum/parser/dag"
    "chronum/engine"
)

func main() {
    // --- 1. Load and parse YAML ---
    content, err := os.ReadFile("examples/chronum.yaml")
    if err != nil {
        panic(fmt.Errorf("failed to read yaml: %w", err))
    }

    src := types.ChronumSource{
        Type:    "yaml",
        Content: content,
    }

    p, err := parser.Get(src.Type)
    if err != nil {
        panic(fmt.Errorf("parser not found: %w", err))
    }

    flow, err := p.Parse(src)
    if err != nil {
        panic(fmt.Errorf("parse error: %w", err))
    }

    fmt.Println("✅ Parsed Chronum:")
    fmt.Printf("Name: %s\n", flow.Name)
    fmt.Println("Steps:")
    for _, s := range flow.Steps {
        fmt.Printf(" - %s (run: %s, needs: %v)\n", s.Name, s.Run, s.Needs)
    }

    // --- 2. Build DAG from steps ---
    d := dag.New()
    for _, step := range flow.Steps {
        d.AddNode(step.Name)
    }

    deps := make(map[string][]string)
    for _, step := range flow.Steps {
        deps[step.Name] = step.Needs
    }

    if err := d.LinkDependencies(deps); err != nil {
        panic(fmt.Errorf("failed to link dependencies: %w", err))
    }

    // --- 3. Create engine components ---
    ctx := engine.NewRunContext(map[string]string{"FLOW_NAME": flow.Name})
    state := engine.NewStateManager()
    shellExec := &engine.ShellExecutor{}
    runner := engine.NewRunner(state, ctx, shellExec)

	// Parallel Execution
	e := engine.New()
	e.RegisterExecutor("shell", &engine.ShellExecutor{})

	fmt.Println("🚀 Running in parallel mode...")
	if err := e.RunParallel(d); err != nil {
		panic(err)
	}

    // --- 4. Execute DAG ---
    fmt.Println("\n🚀 Starting Chronum run...")
    sorted, err := d.TopologicalSort()
    if err != nil {
        panic(fmt.Errorf("topological sort failed: %w", err))
    }

    for _, node := range sorted {
        runner.RunNode(node)
    }

    // --- 5. Print final states ---
    fmt.Println("\n📊 Final step states:")
    for id, status := range state.Snapshot() {
        fmt.Printf(" - %s: %s\n", id, status)
    }

    fmt.Println("\n🎉 Chronum run complete!")
}
