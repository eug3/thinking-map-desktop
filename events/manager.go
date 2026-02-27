/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/events/manager.go
 */
package events

import (
	"context"
	"sync"
)

// EventEmitFn is a function type for emitting events to the frontend
type EventEmitFn func(name string, data ...interface{})

// Event types - matching original SSE events
const (
	EventNodeCreated             = "nodeCreated"
	EventNodeUpdated             = "nodeUpdated"
	EventNodeDeleted             = "nodeDeleted"
	EventNodeDependenciesUpdated = "nodeDependenciesUpdated"
	EventMessageText             = "messageText"
	EventMessageConclusion       = "messageConclusion"
	EventMessageThought          = "messageThought"
	EventMessageAction           = "messageAction"
	EventConclusionCompleted     = "conclusionCompleted"
	EventDecompositionCompleted  = "decompositionCompleted"
	EventError                   = "error"
)

// EventData represents a structured event payload
type EventData struct {
	Type      string                 `json:"type"`
	NodeID    string                 `json:"nodeId,omitempty"`
	MapID     string                 `json:"mapId,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

// EventManager manages event emission
type EventManager struct {
	emitFn    EventEmitFn
	mutex     sync.RWMutex
	listeners map[string][]chan map[string]interface{}
}

// NewEventManager creates a new event manager
func NewEventManager() *EventManager {
	return &EventManager{
		listeners: make(map[string][]chan map[string]interface{}),
	}
}

// SetEmitFn sets the event emission function
// This should be called from main.go after the app is created
func (m *EventManager) SetEmitFn(fn EventEmitFn) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.emitFn = fn
}

// Emit sends an event to the frontend
func (m *EventManager) Emit(eventType string, data interface{}) error {
	m.mutex.RLock()
	emitFn := m.emitFn
	listeners := m.listeners[eventType]
	m.mutex.RUnlock()

	// Send via local emit function if set
	if emitFn != nil {
		emitFn(eventType, data)
	} else {
		// Fall back to global emitter
		EmitGlobal(eventType, data)
	}

	// Also send to local listeners (for Go-side subscribers)
	for _, ch := range listeners {
		select {
		case ch <- map[string]interface{}{
			"type": eventType,
			"data": data,
		}:
		default:
			// Channel full, skip
		}
	}

	return nil
}

// EmitNodeEvent emits a node-related event
func (m *EventManager) EmitNodeEvent(eventType, nodeID, mapID string, data map[string]interface{}) error {
	return m.Emit(eventType, EventData{
		Type:      eventType,
		NodeID:    nodeID,
		MapID:     mapID,
		Data:      data,
		Timestamp: 0, // Will be set by frontend
	})
}

// Subscribe allows internal Go code to listen for events
func (m *EventManager) Subscribe(ctx context.Context, eventType string) <-chan map[string]interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ch := make(chan map[string]interface{}, 100)
	m.listeners[eventType] = append(m.listeners[eventType], ch)

	go func() {
		<-ctx.Done()
		m.Unsubscribe(eventType, ch)
	}()

	return ch
}

// Unsubscribe removes a listener
func (m *EventManager) Unsubscribe(eventType string, ch chan map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	listeners := m.listeners[eventType]
	for i, listener := range listeners {
		if listener == ch {
			m.listeners[eventType] = append(listeners[:i], listeners[i+1:]...)
			close(ch)
			break
		}
	}
}
