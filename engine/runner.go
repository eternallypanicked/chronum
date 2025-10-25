package engine

import (
    "chronum/parser/dag"
    "fmt"
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

func (r *Runner) RunNodeWithRetry(node *dag.Node, retries int) StepResult {
    id := node.ID
    r.State.Set(id, StepRunning)
    fmt.Printf("[%s] Running...\n", id)

    var lastErr error
    for attempt := 1; attempt <= retries+1; attempt++ {
        err := r.Exec.Execute(node)
        if err == nil {
            r.State.Set(id, StepSuccess)
            fmt.Printf("[%s] complete\n", id)
            return StepResult{ID: id, Status: StepSuccess}
        }

        fmt.Printf("[%s] failed (attempt %d/%d): %v\n", id, attempt, retries+1, err)
        lastErr = err
    }

    r.State.Set(id, StepFailed)
    return StepResult{ID: id, Status: StepFailed, Error: lastErr}
}

