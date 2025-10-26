package main

import (
    "fmt"
    "os"

    "chronum/parser"
    "chronum/parser/types"
    _ "chronum/parser/yaml"
    "chronum/parser/dag"
    "chronum/engine"
)

func main() {
    // --- 1. Load YAML ---
    content, err := os.ReadFile("examples/chronum.yaml")
    if err != nil {
        panic(fmt.Errorf("failed to read yaml: %w", err))
    }

    src := types.ChronumSource{Type: "yaml", Content: content}
    p, _ := parser.Get(src.Type)
    flow, err := p.Parse(src)
    if err != nil {
        panic(err)
    }

    fmt.Printf("✅ Parsed Chronum: %s\n", flow.Name)
    for _, s := range flow.Steps {
        fmt.Printf(" - %s (run: %s, needs: %v)\n", s.Name, s.Run, s.Needs)
    }

    // --- 2. Build DAG ---
    d := dag.New()
    for i := range flow.Steps {
        step := &flow.Steps[i]
        node := d.AddNode(step.Name)
        node.Step = step
    }
    deps := make(map[string][]string)
    for _, step := range flow.Steps {
        deps[step.Name] = step.Needs
    }
    if err := d.LinkDependencies(deps); err != nil {
        panic(fmt.Errorf("link dependencies: %w", err))
    }

    // --- 3. Create Engine ---
    e := engine.New()
    e.RegisterExecutor("shell", &engine.ShellExecutor{})
    e.RegisterExecutor("python", &engine.PythonExecutor{})

    // --- 4. Run DAG ---
    fmt.Println("🚀 Running in parallel mode...")
    if err := e.RunParallelWithLimit(d, flow); err != nil {
        panic(err)
    }

    fmt.Println("\n🎉 Chronum run complete!")
}
