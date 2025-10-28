package engine

import (
	"context"
	"sync"
	"time"
)

type RunContext struct {
    ID        string            // unique run ID
    StartTime time.Time         // when it began
    Env       map[string]string // global environment vars
    ctx       context.Context   // standard Go context for cancellation
    cancel    context.CancelFunc
	FlowName  string
	mu        sync.RWMutex
}

func NewRunContext(env map[string]string, flowName string) *RunContext {

	if env == nil {
		env = map[string]string{}
	}
    ctx, cancel := context.WithCancel(context.Background())
    return &RunContext{
        ID:        generateRunID(),
        StartTime: time.Now(),
        Env:       env,
        ctx:       ctx,
        cancel:    cancel,
		FlowName: flowName,
    }
}

func (r *RunContext) Cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *RunContext) IsCancelled() bool {
	select {
	case <-r.ctx.Done():
		return true
	default:
		return false
	}
}

// for now just a timestamp-based run ID
func generateRunID() string {
    return time.Now().Format("20060102T150405")
}

func (r *RunContext) Context() context.Context {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctx
}