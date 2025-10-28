package engine

import (
	"chronum/parser/dag"
	"chronum/parser/types"
	"fmt"
	"sync"
)

// Engine coordinates execution of DAG nodes using registered executors.
type Engine struct {
	mu        sync.RWMutex
	executors map[string]Executor // executor registry
	bus       *SafeEventBus
	// a simple run registry to track running flows (optional introspection)
	runsMu sync.RWMutex
	runs   map[string]*RunContext
}

// New creates a new engine with no executors yet.
func New() *Engine {
	return &Engine{
		executors: make(map[string]Executor),
		bus:       NewEventBus(),
		runs:      make(map[string]*RunContext),
	}
}

// RegisterExecutor adds a new executor type to the engine.
func (e *Engine) RegisterExecutor(name string, ex Executor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.executors[name] = ex
}

func (e *Engine) getExecutor(name string) Executor {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.executors[name]
}

func (e *Engine) Bus() *SafeEventBus  {
	return e.bus
}

func (e *Engine) registerRun(ctx *RunContext) {
	e.runsMu.Lock()
	defer e.runsMu.Unlock()
	e.runs[ctx.ID] = ctx
}

func (e *Engine) unregisterRun(id string) {
	e.runsMu.Lock()
	defer e.runsMu.Unlock()
	delete(e.runs, id)
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

		ex := e.getExecutor("shell")
		if ex == nil {
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
	ctx := NewRunContext(nil, flow.Name)
	// register run for observability
	e.registerRun(ctx)
	defer e.unregisterRun(ctx.ID)

	// choose default executor (shell) but runner will use per-node executor as needed
	defaultExec := e.getExecutor("shell")
	runner := NewRunner(state, ctx, defaultExec, e.Bus())

	fmt.Printf("\nRunning %s (maxParallel=%d, stopOnFail=%v, retry=%d)\n",
		flow.Name, maxParallel, flow.StopOnFail, flow.DefaultRetry)

	// gather root nodes
	ready := []*dag.Node{}
	for _, n := range graph.Nodes {
		if len(n.Dependencies) == 0 {
			ready = append(ready, n)
		}
	}

	// if no nodes, return immediately
	if len(ready) == 0 {
		return nil
	}

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

		go func(n *dag.Node) {
			defer func() {
				<-sem
				wg.Done()
			}()

			// check cancellation early
			if ctx.IsCancelled() {
				return
			}

			step := n.Step
			execType := "shell"
			if step != nil && step.Executor != "" {
				execType = step.Executor
			}

			ex := e.getExecutor(execType)
			if ex == nil {
				// fallback to shell executor if missing
				ex = defaultExec
			}

			// Execute with retry logic
			res := runner.RunNodeWithRetry(n, ex, flow.DefaultRetry)

			// Record step result
			if res.Status == StepFailed {
				state.Set(n.ID, StepFailed)
				fmt.Printf("Step '%s' failed.\n", n.ID)
			} else {
				state.Set(n.ID, StepSuccess)
			}

			// Stop all if configured and a failure occurred
			if res.Status == StepFailed && flow.StopOnFail {
				fmt.Printf("Stop-on-fail active — cancelling remaining steps.\n")
				ctx.Cancel()
				return
			}

			// Launch dependents if all dependencies succeeded
			for _, dep := range n.Dependents {
				if ctx.IsCancelled() {
					return
				}
				if allDepsSatisfied(dep, state) {
					runNode(dep)
				}
			}
		}(node)
	}

	// start all roots
	for _, node := range ready {
		runNode(node)
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
