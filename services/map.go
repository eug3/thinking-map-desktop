/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/services/map.go
 */
package services

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/repositories"
	"github.com/google/uuid"
)

// MapService handles map business logic
type MapService struct {
	mapRepo repositories.ThinkingMapRepository
}

// NewMapService creates a new map service
func NewMapService(mapRepo repositories.ThinkingMapRepository) *MapService {
	return &MapService{mapRepo: mapRepo}
}

// CreateMapRequest represents a request to create a map
type CreateMapRequest struct {
	Title       string   `json:"title"`
	Problem     string   `json:"problem"`
	ProblemType string   `json:"problemType"`
	Target      string   `json:"target"`
	KeyPoints   []string `json:"keyPoints"`
	Constraints []string `json:"constraints"`
}

// UpdateMapRequest represents a request to update a map
type UpdateMapRequest struct {
	Status      int      `json:"status,omitempty"`
	Problem     string   `json:"problem,omitempty"`
	ProblemType string   `json:"problemType,omitempty"`
	Target      string   `json:"target,omitempty"`
	KeyPoints   []string `json:"keyPoints,omitempty"`
	Constraints []string `json:"constraints,omitempty"`
	Conclusion  string   `json:"conclusion,omitempty"`
}

// MapListQuery represents query parameters for listing maps
type MapListQuery struct {
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	Status      int    `json:"status,omitempty"`
	ProblemType string `json:"problemType,omitempty"`
	Search      string `json:"search,omitempty"`
	DateRange   string `json:"dateRange,omitempty"`
}

// MapListResponse represents a paginated list of maps
type MapListResponse struct {
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
	Items []*models.ThinkingMap `json:"items"`
}

// CreateMap creates a new thinking map
func (s *MapService) CreateMap(ctx context.Context, req CreateMapRequest) (*models.ThinkingMap, error) {
	mapID := uuid.NewString()
	thinkingMap := &models.ThinkingMap{
		ID:          mapID,
		Title:       req.Title,
		Problem:     req.Problem,
		ProblemType: req.ProblemType,
		Target:      req.Target,
		KeyPoints:   req.KeyPoints,
		Constraints: req.Constraints,
		Conclusion:  "",
		Metadata:    make(models.Metadata),
		Status:      "initial",
	}

	rootNodeID := uuid.NewString()
	rootNode := &models.ThinkingNode{
		ID:        rootNodeID,
		MapID:     mapID,
		ParentID:  uuid.Nil.String(),
		NodeType:  "root",
		Question:  req.Problem,
		Target:    req.Target,
		Status:    "pending",
		Position:  models.Position{X: 0, Y: 0},
		Context:   models.DependentContext{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.mapRepo.Create(ctx, thinkingMap, rootNode); err != nil {
		return nil, err
	}

	return thinkingMap, nil
}

// GetMap retrieves a specific thinking map
func (s *MapService) GetMap(ctx context.Context, mapID string) (*models.ThinkingMap, error) {
	return s.mapRepo.FindByID(ctx, mapID)
}

// ListMaps retrieves a list of thinking maps
func (s *MapService) ListMaps(ctx context.Context, query MapListQuery) (*MapListResponse, error) {
	// Get current time in UTC+8 (China Standard Time)
	now := time.Now().In(time.FixedZone("CST", 8*60*60))
	var startTime, endTime time.Time
	if query.DateRange != "" {
		switch query.DateRange {
		case "this-week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			endTime = now
		case "last-week":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startTime = now.AddDate(0, 0, -weekday-6).Truncate(24 * time.Hour)
			endTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
		case "this-month":
			startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.FixedZone("CST", 8*60*60))
			endTime = now
		case "all-time":
			break
		}
	}

	status := ""
	if query.Status > 0 {
		statusMap := map[int]string{
			1: "initial",
			2: "running",
			3: "completed",
			4: "deleted",
		}
		status = statusMap[query.Status]
	}

	maps, total, err := s.mapRepo.List(ctx, status, query.ProblemType, query.Search, startTime, endTime, query.Page, query.Limit)
	if err != nil {
		return nil, err
	}

	return &MapListResponse{
		Total: int(total),
		Page:  query.Page,
		Limit: query.Limit,
		Items: maps,
	}, nil
}

// UpdateMap updates a thinking map
func (s *MapService) UpdateMap(ctx context.Context, mapID string, req UpdateMapRequest) (*models.ThinkingMap, error) {
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if req.Status > 0 {
		statusMap := map[int]string{
			1: "initial",
			2: "running",
			3: "completed",
			4: "deleted",
		}
		if status, ok := statusMap[req.Status]; ok {
			updates["status"] = status
		}
	}
	if req.Problem != "" {
		updates["problem"] = req.Problem
	}
	if req.ProblemType != "" {
		updates["problem_type"] = req.ProblemType
	}
	if req.Target != "" {
		updates["target"] = req.Target
	}
	if req.KeyPoints != nil {
		updates["key_points"] = req.KeyPoints
	}
	if req.Constraints != nil {
		updates["constraints"] = req.Constraints
	}
	if req.Conclusion != "" {
		updates["conclusion"] = req.Conclusion
	}
	if err := s.mapRepo.Update(ctx, mapID, updates); err != nil {
		return nil, err
	}
	return s.mapRepo.FindByID(ctx, mapID)
}

// DeleteMap deletes a thinking map
func (s *MapService) DeleteMap(ctx context.Context, mapID string) error {
	return s.mapRepo.Delete(ctx, mapID)
}
