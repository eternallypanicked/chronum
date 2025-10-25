package engine

import (
    "context"
    "time"
)

type RunContext struct {
    ID        string            // unique run ID
    StartTime time.Time         // when it began
    Env       map[string]string // global environment vars
    Ctx       context.Context   // standard Go context for cancellation
    Cancel    context.CancelFunc
}

func NewRunContext(env map[string]string) *RunContext {
    ctx, cancel := context.WithCancel(context.Background())
    return &RunContext{
        ID:        generateRunID(),
        StartTime: time.Now(),
        Env:       env,
        Ctx:       ctx,
        Cancel:    cancel,
    }
}

// for now just a timestamp-based run ID
func generateRunID() string {
    return time.Now().Format("20060102T150405")
}
