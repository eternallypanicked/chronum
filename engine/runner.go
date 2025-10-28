package engine

import (
	"chronum/parser/dag"
	"fmt"
	"math/rand"
	"time"
)

type Runner struct {
    State *StateManager
    ctx   *RunContext
    Exec  Executor
	Emitter EventEmitter
}

func NewRunner(state *StateManager, ctx *RunContext, exec Executor, emitter EventEmitter) *Runner {
	return &Runner{
		State:   state,
		ctx:     ctx,
		Exec:    exec,
		Emitter: emitter,
	}
}

func (r *Runner) RunNode(node *dag.Node) StepResult {
	id := node.ID
	r.State.Set(id, StepRunning)
	fmt.Printf("[%s] Running...\n", id)

	// Basic cancellation check
	if r.ctx != nil && r.ctx.IsCancelled() {
		r.State.Set(id, StepFailed)
		return StepResult{ID: id, Status: StepFailed, Error: fmt.Errorf("cancelled")}
	}

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

	// emit start
	if r.Emitter != nil {
		r.Emitter.Emit(Event{
			Type:      EventStepStart,
			Timestamp: time.Now().UTC(),
			FlowName:  r.ctx.FlowName,
			StepName:  id,
			Status:    StepRunning,
			Attempt:   1,
			MaxRetry:  retries + 1,
		})
	}

	var lastErr error
	delay := time.Second

	for attempt := 1; attempt <= retries+1; attempt++ {
		if r.ctx != nil && r.ctx.IsCancelled() {
			r.State.Set(id, StepFailed)
			if r.Emitter != nil {
				r.Emitter.Emit(Event{
					Type:      EventStepFail,
					Timestamp: time.Now().UTC(),
					FlowName:  r.ctx.FlowName,
					StepName:  id,
					Status:    StepError,
					Attempt:   attempt,
					MaxRetry:  retries + 1,
					Error:     "cancelled",
				})
			}
			return StepResult{ID: id, Status: StepFailed, Error: fmt.Errorf("cancelled")}
		}

		err := ex.Execute(node)
		if err == nil {
			r.State.Set(id, StepSuccess)
			if r.Emitter != nil {
				r.Emitter.Emit(Event{
					Type:      EventStepSuccess,
					Timestamp: time.Now().UTC(),
					FlowName:  r.ctx.FlowName,
					StepName:  id,
					Status:    StepOK,
					Attempt:   attempt,
					MaxRetry:  retries + 1,
				})
			}
			return StepResult{ID: id, Status: StepSuccess}
		}

		// record failure for this attempt
		lastErr = err
		if r.Emitter != nil {
			r.Emitter.Emit(Event{
				Type:      EventStepFail,
				Timestamp: time.Now().UTC(),
				FlowName:  r.ctx.FlowName,
				StepName:  id,
				Status:    StepError,
				Attempt:   attempt,
				MaxRetry:  retries + 1,
				Error:     err.Error(),
			})
		}

		if attempt < retries+1 {
			// exponential backoff with jitter
			jitter := time.Duration(rand.Intn(250)) * time.Millisecond
			time.Sleep(delay + jitter)
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}

	// out of retries
	r.State.Set(id, StepFailed)
	return StepResult{ID: id, Status: StepFailed, Error: lastErr}
}

