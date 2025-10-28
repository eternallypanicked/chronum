package main

import (
	"fmt"
	"os"
	"time"

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

	fmt.Printf("Parsed Chronum: %s\n", flow.Name)
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
	e.RegisterExecutor("shell", &engine.ShellExecutor{Timeout: 5 * time.Minute})
	e.RegisterExecutor("python", &engine.PythonExecutor{Timeout: 5 * time.Minute})

	// --- 4. Subscribe to EventBus ---
	e.Bus().Subscribe(func(ev engine.Event) {
		ts := ev.Timestamp.Format(time.StampMilli)
		switch ev.Type {
		case engine.EventStepStart:
			fmt.Printf("[event] %s → step.start: %s\n", ts, ev.StepName)
		case engine.EventStepSuccess:
			fmt.Printf("[event] %s → step.success: %s ✅\n", ts, ev.StepName)
		case engine.EventStepFail:
			fmt.Printf("[event] %s → step.fail: %s ❌ (%s)\n", ts, ev.StepName, ev.Error)
		case engine.EventFlowStart:
			fmt.Printf("[event] %s → flow.start: %s\n", ts, ev.FlowName)
		case engine.EventFlowComplete:
			fmt.Printf("[event] %s → flow.complete: %s (%s)\n", ts, ev.FlowName, ev.Status)
		}
	})

	// --- 5. Run DAG ---
	fmt.Println("Running in parallel mode...")
	if err := e.RunParallelWithLimit(d, flow); err != nil {
		fmt.Printf("\nFlow ended with errors: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nChronum run complete!")
}
