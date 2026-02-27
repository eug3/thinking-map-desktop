/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/services/node.go
 */
package services

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/repositories"
	"github.com/google/uuid"
)

// NodeService handles node business logic
type NodeService struct {
	nodeRepo repositories.ThinkingNodeRepository
	mapRepo  repositories.ThinkingMapRepository
}

// NewNodeService creates a new node service
func NewNodeService(nodeRepo repositories.ThinkingNodeRepository, mapRepo repositories.ThinkingMapRepository) *NodeService {
	return &NodeService{
		nodeRepo: nodeRepo,
		mapRepo:  mapRepo,
	}
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

// ExecutableNodesResponse represents executable nodes
type ExecutableNodesResponse struct {
	NodeIDs         []string `json:"nodeIds"`
	SuggestedNodeID string   `json:"suggestedNodeId"`
}

// ListNodes retrieves all nodes in a map
func (s *NodeService) ListNodes(ctx context.Context, mapID string) ([]*models.ThinkingNode, error) {
	nodes, err := s.nodeRepo.ByMapID(ctx, mapID)
	if err != nil {
		return nil, err
	}
	// Populate context for each node
	for _, node := range nodes {
		node.Context = s.getNodeContext(ctx, node)
	}
	return nodes, nil
}

// CreateNode creates a new node
func (s *NodeService) CreateNode(ctx context.Context, mapID string, req CreateNodeRequest) (*models.ThinkingNode, error) {
	node := &models.ThinkingNode{
		ID:       uuid.NewString(),
		MapID:    mapID,
		ParentID: req.ParentID,
		NodeType: req.NodeType,
		Question: req.Question,
		Target:   req.Target,
		Status:   "pending",
		Position: req.Position,
		Decomposition: models.Decomposition{
			IsDecomposed:   false,
			LastMessageID:  "",
			ConversationID: "",
		},
		Conclusion: models.Conclusion{
			LastMessageID: "",
			Content:       "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.nodeRepo.Create(ctx, node); err != nil {
		return nil, err
	}
	return node, nil
}

// UpdateNode updates a node
func (s *NodeService) UpdateNode(ctx context.Context, nodeID string, req UpdateNodeRequest) (*models.ThinkingNode, error) {
	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	if req.Question != "" {
		node.Question = req.Question
	}
	if req.Target != "" {
		node.Target = req.Target
	}
	if (req.Position != models.Position{}) {
		node.Position = req.Position
	}
	if req.Status != "" {
		node.Status = req.Status
	}
	node.UpdatedAt = time.Now()
	if err := s.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}
	return node, nil
}

// DeleteNode deletes a node
func (s *NodeService) DeleteNode(ctx context.Context, nodeID string) error {
	return s.nodeRepo.Delete(ctx, nodeID)
}

// ExecutableNodes retrieves executable nodes
func (s *NodeService) ExecutableNodes(ctx context.Context, mapID, nodeID string) (*ExecutableNodesResponse, error) {
	nodes, err := s.nodeRepo.ByMapID(ctx, mapID)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]*models.ThinkingNode)
	for _, node := range nodes {
		nodeMap[node.ID] = node
	}

	childrenMap := make(map[string][]*models.ThinkingNode)
	for _, node := range nodes {
		if node.ParentID != "" && node.ParentID != uuid.Nil.String() {
			childrenMap[node.ParentID] = append(childrenMap[node.ParentID], node)
		}
	}

	var executableNodeIDs []string
	var nodesToUpdate []*models.ThinkingNode

	// Process node status transitions
	for _, node := range nodes {
		if node.Status == "initial" {
			canExecute := true
			for _, depID := range node.Dependencies {
				depNode, exists := nodeMap[depID]
				if !exists {
					continue
				}
				if depNode.Status != "completed" {
					canExecute = false
					break
				}
			}

			if canExecute {
				node.Status = "pending"
				nodesToUpdate = append(nodesToUpdate, node)
			}
		}

		if node.Status == "in_decomposition" {
			allChildrenCompleted := true
			childNodes := childrenMap[node.ID]
			if len(childNodes) > 0 {
				for _, childNode := range childNodes {
					if childNode.Status != "completed" {
						allChildrenCompleted = false
						break
					}
				}

				if allChildrenCompleted {
					node.Status = "in_conclusion"
					nodesToUpdate = append(nodesToUpdate, node)
				}
			}
		}
	}

	for _, node := range nodes {
		if node.Status == "pending" {
			executableNodeIDs = append(executableNodeIDs, node.ID)
		}
	}

	for _, node := range nodesToUpdate {
		if err := s.nodeRepo.Update(ctx, node); err != nil {
			return nil, err
		}
	}

	var suggestedNodeID string
	if nodeID != "" {
		currentNode, exists := nodeMap[nodeID]
		if exists {
			switch currentNode.Status {
			case "completed":
				if currentNode.ParentID != "" && currentNode.ParentID != uuid.Nil.String() {
					parentNode, parentExists := nodeMap[currentNode.ParentID]
					if parentExists {
						siblings := childrenMap[parentNode.ID]
						for _, sibling := range siblings {
							if sibling.ID != currentNode.ID && sibling.Status == "pending" {
								suggestedNodeID = sibling.ID
								break
							}
						}
						if suggestedNodeID == "" && parentNode.Status == "pending" {
							suggestedNodeID = parentNode.ID
						}
					}
				}
				if suggestedNodeID == "" && len(executableNodeIDs) > 0 {
					suggestedNodeID = executableNodeIDs[0]
				}
			case "in_conclusion":
				suggestedNodeID = nodeID
			default:
				childNodes := childrenMap[currentNode.ID]
				for _, childNode := range childNodes {
					if childNode.Status == "pending" {
						suggestedNodeID = childNode.ID
						break
					}
				}
				if suggestedNodeID == "" && currentNode.Status == "pending" {
					suggestedNodeID = nodeID
				} else if suggestedNodeID == "" && len(executableNodeIDs) > 0 {
					suggestedNodeID = executableNodeIDs[0]
				}
			}
		}
	} else if len(executableNodeIDs) > 0 {
		suggestedNodeID = executableNodeIDs[0]
	}

	return &ExecutableNodesResponse{
		NodeIDs:         executableNodeIDs,
		SuggestedNodeID: suggestedNodeID,
	}, nil
}

// getNodeContext retrieves the node context
func (s *NodeService) getNodeContext(ctx context.Context, node *models.ThinkingNode) models.DependentContext {
	ancestor := s.getAncestor(ctx, node.ID)
	prevSibling := s.getPreSibling(ctx, node)
	children := s.getChildren(ctx, node.ID)

	return models.DependentContext{
		Ancestor:    ancestor,
		PrevSibling: prevSibling,
		Children:    children,
	}
}

// getAncestor recursively gets all ancestor nodes
func (s *NodeService) getAncestor(ctx context.Context, nodeID string) []models.NodeContext {
	var nodeContexts []models.NodeContext

	node, err := s.nodeRepo.FindByID(ctx, nodeID)
	if err != nil || node.ParentID == "" {
		return nodeContexts
	}

	parentNode, err := s.nodeRepo.FindByID(ctx, node.ParentID)
	if err != nil {
		return nodeContexts
	}

	nodeContext := models.NodeContext{
		Question: parentNode.Question,
		Target:   parentNode.Target,
		Abstract: "",
		Status:   parentNode.Status,
	}
	ancestor := s.getAncestor(ctx, parentNode.ID)
	ancestor = append(ancestor, nodeContext)

	return ancestor
}

// getPreSibling gets all previous sibling nodes
func (s *NodeService) getPreSibling(ctx context.Context, node *models.ThinkingNode) []models.NodeContext {
	var nodeContexts []models.NodeContext

	if len(node.Dependencies) == 0 {
		return nodeContexts
	}

	depNodes, err := s.nodeRepo.FindByIDs(ctx, node.Dependencies)
	if err != nil {
		return nodeContexts
	}

	for _, depNode := range depNodes {
		nodeContext := models.NodeContext{
			Question:   depNode.Question,
			Target:     depNode.Target,
			Conclusion: depNode.Conclusion.Content,
			Abstract:   "",
			Status:     depNode.Status,
		}
		nodeContexts = append(nodeContexts, nodeContext)
	}

	return nodeContexts
}

// getChildren gets all direct child nodes
func (s *NodeService) getChildren(ctx context.Context, nodeID string) []models.NodeContext {
	var nodeContexts []models.NodeContext

	childNodes, err := s.nodeRepo.ByParentID(ctx, nodeID)
	if err != nil {
		return nodeContexts
	}

	for _, childNode := range childNodes {
		nodeContext := models.NodeContext{
			Question:   childNode.Question,
			Target:     childNode.Target,
			Conclusion: childNode.Conclusion.Content,
			Abstract:   "",
			Status:     childNode.Status,
		}
		nodeContexts = append(nodeContexts, nodeContext)
	}

	return nodeContexts
}
