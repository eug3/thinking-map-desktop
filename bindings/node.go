/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/bindings/node.go
 */
package bindings

import (
	"context"

	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/services"
)

// NodeBinding exposes node operations to the frontend
type NodeBinding struct {
	service *services.NodeService
}

// NewNodeBinding creates a new node binding
func NewNodeBinding(service *services.NodeService) *NodeBinding {
	return &NodeBinding{service: service}
}

// CreateNodeRequest represents a request to create a node
type CreateNodeRequest struct {
	ParentID string          `json:"parentId"`
	NodeType string          `json:"nodeType"`
	Question string          `json:"question"`
	Target   string          `json:"target"`
	Position models.Position `json:"position"`
}

// UpdateNodeRequest represents a request to update a node
type UpdateNodeRequest struct {
	Question string          `json:"question,omitempty"`
	Target   string          `json:"target,omitempty"`
	Position models.Position `json:"position,omitempty"`
	Status   string          `json:"status,omitempty"`
}

// NodeResponse represents a node in API responses
type NodeResponse struct {
	ID            string                  `json:"id"`
	MapID         string                  `json:"mapId"`
	ParentID      string                  `json:"parentId"`
	NodeType      string                  `json:"nodeType"`
	Question      string                  `json:"question"`
	Target        string                  `json:"target"`
	Context       models.DependentContext `json:"context"`
	Decomposition models.Decomposition    `json:"decomposition"`
	Conclusion    models.Conclusion       `json:"conclusion"`
	Status        string                  `json:"status"`
	Position      models.Position         `json:"position"`
	Dependencies  []string                `json:"dependencies"`
	CreatedAt     int64                   `json:"createdAt"`
	UpdatedAt     int64                   `json:"updatedAt"`
}

// ExecutableNodesResponse represents executable nodes
type ExecutableNodesResponse struct {
	NodeIDs         []string `json:"nodeIds"`
	SuggestedNodeID string   `json:"suggestedNodeId"`
}

// ListNodes retrieves all nodes in a map
func (b *NodeBinding) ListNodes(ctx context.Context, mapID string) ([]NodeResponse, error) {
	nodes, err := b.service.ListNodes(ctx, mapID)
	if err != nil {
		return nil, err
	}

	responses := make([]NodeResponse, len(nodes))
	for i, n := range nodes {
		responses[i] = toNodeResponse(n)
	}
	return responses, nil
}

// CreateNode creates a new node
func (b *NodeBinding) CreateNode(ctx context.Context, mapID string, req CreateNodeRequest) (NodeResponse, error) {
	node, err := b.service.CreateNode(ctx, mapID, services.CreateNodeRequest{
		ParentID: req.ParentID,
		NodeType: req.NodeType,
		Question: req.Question,
		Target:   req.Target,
		Position: req.Position,
	})
	if err != nil {
		return NodeResponse{}, err
	}
	return toNodeResponse(node), nil
}

// UpdateNode updates a node
func (b *NodeBinding) UpdateNode(ctx context.Context, nodeID string, req UpdateNodeRequest) (NodeResponse, error) {
	node, err := b.service.UpdateNode(ctx, nodeID, services.UpdateNodeRequest{
		Question: req.Question,
		Target:   req.Target,
		Position: req.Position,
		Status:   req.Status,
	})
	if err != nil {
		return NodeResponse{}, err
	}
	return toNodeResponse(node), nil
}

// DeleteNode deletes a node
func (b *NodeBinding) DeleteNode(ctx context.Context, nodeID string) error {
	return b.service.DeleteNode(ctx, nodeID)
}

// ExecutableNodes retrieves executable nodes
func (b *NodeBinding) ExecutableNodes(ctx context.Context, mapID, nodeID string) (ExecutableNodesResponse, error) {
	result, err := b.service.ExecutableNodes(ctx, mapID, nodeID)
	if err != nil {
		return ExecutableNodesResponse{}, err
	}
	return ExecutableNodesResponse{
		NodeIDs:         result.NodeIDs,
		SuggestedNodeID: result.SuggestedNodeID,
	}, nil
}

// InitNodeBinding initializes a NodeBinding with its service dependency
func InitNodeBinding(b *NodeBinding, service *services.NodeService) {
	b.service = service
}

func toNodeResponse(n *models.ThinkingNode) NodeResponse {
	return NodeResponse{
		ID:            n.ID,
		MapID:         n.MapID,
		ParentID:      n.ParentID,
		NodeType:      n.NodeType,
		Question:      n.Question,
		Target:        n.Target,
		Context:       n.Context,
		Decomposition: n.Decomposition,
		Conclusion:    n.Conclusion,
		Status:        n.Status,
		Position:      n.Position,
		Dependencies:  n.Dependencies,
		CreatedAt:     n.CreatedAt.Unix(),
		UpdatedAt:     n.UpdatedAt.Unix(),
	}
}
