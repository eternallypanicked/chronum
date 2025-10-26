package engine

import (
	"chronum/parser/dag"
	"chronum/parser/types"
	"fmt"
	"sync"
)

// Engine coordinates execution of DAG nodes using registered executors.
type Engine struct {
	executors map[string]Executor // executor registry
}

// New creates a new engine with no executors yet.
func New() *Engine {
	return &Engine{
		executors: make(map[string]Executor),
	}
}

// RegisterExecutor adds a new executor type to the engine.
func (e *Engine) RegisterExecutor(name string, ex Executor) {
	e.executors[name] = ex
}

// Run executes the DAG in topological order (parallel when possible).
func (e *Engine) Run(graph *dag.DAG) error {
	sorted, err := graph.TopologicalSort()
	if err != nil {
		return fmt.Errorf("topological sort failed: %w", err)
	}

	var mu sync.Mutex
	statuses := make(map[string]string)

	for _, node := range sorted {
		// Simple: every node runs sequentially for now
		fmt.Printf("Running step: %s\n", node.ID)

		// Choose an executor type (for now, assume "shell")
		ex, ok := e.executors["shell"]
		if !ok {
			return fmt.Errorf("no executor registered for 'shell'")
		}

		err := ex.Execute(node)
		mu.Lock()
		statuses[node.ID] = ifErr(err)
		mu.Unlock()

		if err != nil {
			fmt.Printf("step %s failed: %v\n", node.ID, err)
			return err
		}

		fmt.Printf("step %s done\n", node.ID)
	}

	return nil
}

func (e *Engine) RunParallel(graph *dag.DAG) error {
	var mu sync.Mutex
	wg := &sync.WaitGroup{}
	inProgress := make(map[string]bool)

	// Step 1: find all nodes with no dependencies
	ready := []*dag.Node{}
	for _, n := range graph.Nodes {
		if len(n.Dependencies) == 0 {
			ready = append(ready, n)
		}
	}

	state := NewStateManager()
	ctx := NewRunContext(map[string]string{"RUN_MODE": "parallel"})
	runner := NewRunner(state, ctx, e.executors["shell"])

	// Step 2: run nodes recursively
	var runNode func(*dag.Node)
	runNode = func(node *dag.Node) {
		mu.Lock()
		if inProgress[node.ID] {
			mu.Unlock()
			return
		}
		inProgress[node.ID] = true
		mu.Unlock()

		wg.Add(1)
		go func() {
			defer wg.Done()
			res := runner.RunNode(node)
			if res.Status == StepFailed {
				return // early exit if one fails
			}

			// Check dependents
			for _, dep := range node.Dependents {
				if allDepsSatisfied(dep, state) {
					runNode(dep)
				}
			}
		}()
	}

	// Step 3: start all roots
	for _, node := range ready {
		runNode(node)
	}

	wg.Wait()
	fmt.Println("Parallel run complete")
	return nil
}

func (e *Engine) RunParallelWithLimit(graph *dag.DAG, flow *types.ChronumDefinition) error {

	maxParallel := flow.MaxParallel
	if maxParallel < 1 {
		maxParallel = 1
	}

	sem := make(chan struct{}, maxParallel)
	var mu sync.Mutex
	wg := &sync.WaitGroup{}
	inProgress := make(map[string]bool)

	state := NewStateManager()
	ctx := NewRunContext(nil)
	runner := NewRunner(state, ctx, e.executors["shell"])

	fmt.Printf("\nRunning %s (maxParallel=%d, stopOnFail=%v, retry=%d)\n",
		flow.Name, maxParallel, flow.StopOnFail, flow.DefaultRetry)

	var runNode func(*dag.Node)
	runNode = func(node *dag.Node) {
		mu.Lock()
		if inProgress[node.ID] {
			mu.Unlock()
			return
		}
		inProgress[node.ID] = true
		mu.Unlock()

		sem <- struct{}{}
		wg.Add(1)

		go func() {
			defer func() {
				<-sem
				wg.Done()
			}()

			if ctx.IsCancelled() {
				return
			}

			step := node.Step
			execType := step.Executor
			if execType == "" {
				execType = "shell"
			}

			ex, ok := e.executors[execType]
			if !ok {
				fmt.Printf("No executor registered for '%s', using shell\n", execType)
				ex = e.executors["shell"]
			}

			// Execute with retry logic
			res := runner.RunNodeWithRetry(node, ex, flow.DefaultRetry)

			// Record step result
			if res.Status == StepFailed {
				state.Set(node.ID, StepFailed)
				fmt.Printf("Step '%s' failed.\n", node.ID)
			} else {
				state.Set(node.ID, StepSuccess)
			}

			// Stop all if configured and a failure occurred
			if res.Status == StepFailed && flow.StopOnFail {
				fmt.Printf("Stop-on-fail active — cancelling remaining steps.\n")
				ctx.Cancel()
				return
			}

			// Launch dependents if all dependencies succeeded
			for _, dep := range node.Dependents {
				if ctx.IsCancelled() {
					return
				}
				if allDepsSatisfied(dep, state) {
					runNode(dep)
				}
			}
		}()
	}

	// --- Start all root nodes (no dependencies) ---
	for _, node := range graph.Nodes {
		if len(node.Dependencies) == 0 {
			runNode(node)
		}
	}

	wg.Wait()
	fmt.Println("Flow complete.")

	// Detect any failed steps for overall result
	for id, st := range state.Snapshot() {
		if st == StepFailed {
			return fmt.Errorf("flow completed with failed step(s), e.g. %s", id)
		}
	}
	return nil
}

// helper: check if all dependencies succeeded
func allDepsSatisfied(node *dag.Node, state *StateManager) bool {
	for _, dep := range node.Dependencies {
		if state.Get(dep.ID) != StepSuccess {
			return false
		}
	}
	return true
}

func ifErr(err error) string {
	if err != nil {
		return "failed"
	}
	return "success"
}
