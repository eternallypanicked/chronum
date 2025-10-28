package engine

import (
	"fmt"
	"sync"
	"time"
)

// EventType defines known structured events
type EventType string

const (
	EventStepStart    EventType = "step.start"
	EventStepSuccess  EventType = "step.success"
	EventStepFail     EventType = "step.fail"
	EventFlowStart    EventType = "flow.start"
	EventFlowComplete EventType = "flow.complete"
)

// Event is the sanitized, structured message emitted to subscribers.
type Event struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"ts"`

	FlowName string     `json:"flow,omitempty"`
	StepName string     `json:"step,omitempty"`
	Status   StepStatus `json:"status,omitempty"`

	Attempt  int    `json:"attempt,omitempty"`
	MaxRetry int    `json:"maxRetry,omitempty"`
	Error    string `json:"error,omitempty"`
}

// EventEmitter interface
type EventEmitter interface {
	Emit(Event)
}

// SafeEventBus is a concurrency-safe bus. It preserves the public Subscribe(fn func(Event))
// signature for backward compatibility while protecting the emitter from blocking subscribers,
// panics, and slow consumers.
type SafeEventBus struct {
	mu          sync.RWMutex
	subscribers []chan Event
	closed      bool
}

// NewEventBus constructs a new SafeEventBus.
func NewEventBus() *SafeEventBus {
	return &SafeEventBus{
		subscribers: make([]chan Event, 0, 4),
	}
}

// Subscribe adds a subscriber function; the function runs in a dedicated goroutine and receives events.
// The function is executed in a recover wrapper; slow subscribers won't block Emit thanks to per-subscriber buffers.
func (b *SafeEventBus) Subscribe(fn func(Event)) {
	// create channel per subscriber with moderate buffer to allow short bursts without dropping
	ch := make(chan Event, 128)

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		// if bus is closed, run the subscriber and exit immediately
		go func() {
			defer func() { _ = recover() }()
			return
		}()
		return
	}
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()

	// start goroutine that consumes from the channel and calls fn
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// don't allow subscriber panic to terminate the process
				fmt.Printf("[eventbus] subscriber panicked: %v\n", r)
			}
		}()

		for ev := range ch {
			// call subscriber
			fn(ev)
		}
	}()
}

// Emit publishes an event to all subscribers non-blockingly. If a subscriber is full,
// the event is dropped for that subscriber and a warning is printed.
func (b *SafeEventBus) Emit(ev Event) {
	// Truncate long errors to prevent memory abuse or leakage
	if len(ev.Error) > 512 {
		ev.Error = ev.Error[:512] + "...(truncated)"
	}

	b.mu.RLock()
	subs := make([]chan Event, len(b.subscribers))
	copy(subs, b.subscribers)
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- ev:
			// delivered
		default:
			// subscriber too slow; drop and warn
			fmt.Printf("[eventbus] warning: subscriber slow, dropping event %s for flow %s\n", ev.Type, ev.FlowName)
		}
	}
}

// Close gracefully closes all subscriber channels (no further emits should be called after Close).
func (b *SafeEventBus) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true
	subs := b.subscribers
	b.subscribers = nil
	b.mu.Unlock()

	for _, ch := range subs {
		close(ch)
	}
}
