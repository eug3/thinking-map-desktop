/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/bindings/thinking.go
 */
package bindings

import (
	"context"

	"github.com/PGshen/thinking-map-desktop/events"
	"github.com/PGshen/thinking-map-desktop/services"
)

// ThinkingBinding exposes AI thinking operations to the frontend
type ThinkingBinding struct {
	service      *services.ThinkingService
	eventManager *events.EventManager
}

// NewThinkingBinding creates a new thinking binding
func NewThinkingBinding(service *services.ThinkingService, eventManager *events.EventManager) *ThinkingBinding {
	return &ThinkingBinding{
		service:      service,
		eventManager: eventManager,
	}
}

// UnderstandingRequest represents a request for understanding
type UnderstandingRequest struct {
	NodeID string `json:"nodeId"`
	Query  string `json:"query"`
}

// DecompositionRequest represents a request for decomposition
type DecompositionRequest struct {
	NodeID        string `json:"nodeId"`
	Clarification string `json:"clarification"`
	IsDecomposed  bool   `json:"isDecomposed"`
}

// ConclusionRequest represents a request for conclusion
type ConclusionRequest struct {
	NodeID      string `json:"nodeId"`
	Reference   string `json:"reference"`
	Instruction string `json:"instruction"`
}

// Understanding initiates the understanding process (streaming)
func (b *ThinkingBinding) Understanding(ctx context.Context, req UnderstandingRequest) error {
	return b.service.Understanding(ctx, req.NodeID, req.Query, func(chunk string, eventType string) {
		b.emitEvent(req.NodeID, "", eventType, map[string]interface{}{
			"message": chunk,
		})
	})
}

// Decomposition initiates the decomposition process (streaming)
func (b *ThinkingBinding) Decomposition(ctx context.Context, req DecompositionRequest) error {
	return b.service.Decomposition(ctx, req.NodeID, req.Clarification, req.IsDecomposed, func(chunk string, eventType string) {
		b.emitEvent(req.NodeID, "", eventType, map[string]interface{}{
			"message": chunk,
		})
	})
}

// Conclusion initiates the conclusion process (streaming)
func (b *ThinkingBinding) Conclusion(ctx context.Context, req ConclusionRequest) error {
	return b.service.Conclusion(ctx, req.NodeID, req.Reference, req.Instruction, func(chunk string, eventType string) {
		b.emitEvent(req.NodeID, "", eventType, map[string]interface{}{
			"message": chunk,
		})
	})
}

// Repeat re-runs the thinking process
func (b *ThinkingBinding) Repeat(ctx context.Context, nodeID string) error {
	return b.service.Repeat(ctx, nodeID, func(chunk string, eventType string) {
		b.emitEvent(nodeID, "", eventType, map[string]interface{}{
			"message": chunk,
		})
	})
}

// InitThinkingBinding initializes a ThinkingBinding with its dependencies
func InitThinkingBinding(b *ThinkingBinding, service *services.ThinkingService, eventManager *events.EventManager) {
	b.service = service
	b.eventManager = eventManager
}

func (b *ThinkingBinding) emitEvent(nodeID, mapID, eventType string, data map[string]interface{}) {
	// Emit through the event manager which will forward to the frontend
	b.eventManager.EmitNodeEvent(eventType, nodeID, mapID, data)
}
