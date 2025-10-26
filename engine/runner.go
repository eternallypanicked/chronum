package engine

import (
	"chronum/parser/dag"
	"fmt"
	"math/rand"
	"time"
)

type Runner struct {
    State *StateManager
    Ctx   *RunContext
    Exec  Executor
}

func NewRunner(state *StateManager, ctx *RunContext, exec Executor) *Runner {
    return &Runner{State: state, Ctx: ctx, Exec: exec}
}

func (r *Runner) RunNode(node *dag.Node) StepResult {
    id := node.ID
    r.State.Set(id, StepRunning)
    fmt.Printf("[%s] Running...\n", id)

    err := r.Exec.Execute(node)
    if err != nil {
        r.State.Set(id, StepFailed)
        fmt.Printf("[%s] failed: %v\n", id, err)
        return StepResult{ID: id, Status: StepFailed, Error: err}
    }

    r.State.Set(id, StepSuccess)
    fmt.Printf("[%s] complete\n", id)
    return StepResult{ID: id, Status: StepSuccess}
}

// RunNodeWithRetry executes a single DAG node with retry policy, exponential backoff,
// and optional jitter to avoid retry storms.
func (r *Runner) RunNodeWithRetry(node *dag.Node, ex Executor, retries int) StepResult {
	
	id := node.ID
	r.State.Set(id, StepRunning)
	fmt.Printf("[%s] Running...\n", id)

	var lastErr error
	delay := time.Second // initial backoff duration

	for attempt := 1; attempt <= retries+1; attempt++ {
		// respect context cancellation
		if r.Ctx.IsCancelled() {
			r.State.Set(id, StepFailed)
			fmt.Printf("[%s] cancelled before attempt %d\n", id, attempt)
			return StepResult{ID: id, Status: StepFailed, Error: fmt.Errorf("cancelled")}
		}

		err := ex.Execute(node)
		if err == nil {
			r.State.Set(id, StepSuccess)
			fmt.Printf("[%s] complete\n", id)
			return StepResult{ID: id, Status: StepSuccess}
		}

		// record the failure
		lastErr = err
		fmt.Printf("[%s] failed (attempt %d/%d): %v\n", id, attempt, retries+1, err)

		// if this was the final attempt, break
		if attempt == retries+1 {
			break
		}

		// --- exponential backoff with jitter ---
		// random jitter between 0–250 ms
		jitter := time.Duration(rand.Intn(250)) * time.Millisecond
		fmt.Printf("[%s] retrying in %s (±%s jitter)\n", id, delay, jitter)

		time.Sleep(delay + jitter)
		delay *= 2 
		if delay > 30*time.Second {
			delay = 30 * time.Second 
		}
	}

	r.State.Set(id, StepFailed)
	fmt.Printf("[%s] failed after %d attempts\n", id, retries+1)
	return StepResult{ID: id, Status: StepFailed, Error: lastErr}
}

