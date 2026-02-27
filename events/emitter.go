/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/events/emitter.go
 */
package events

import "sync"

// globalEmitFn is the global event emission function
// It should be set from main.go after the Wails app is created
var globalEmitFn EventEmitFn
var emitFnMutex sync.RWMutex

// SetGlobalEmitFn sets the global event emission function
func SetGlobalEmitFn(fn EventEmitFn) {
	emitFnMutex.Lock()
	defer emitFnMutex.Unlock()
	globalEmitFn = fn
}

// EmitGlobal sends an event to the frontend using the global emitter
func EmitGlobal(eventType string, data interface{}) {
	emitFnMutex.RLock()
	fn := globalEmitFn
	emitFnMutex.RUnlock()

	if fn != nil {
		fn(eventType, data)
	}
}
