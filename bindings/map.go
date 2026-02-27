/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/bindings/map.go
 */
package bindings

import (
	"context"

	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/services"
)

// MapBinding exposes map operations to the frontend
type MapBinding struct {
	service *services.MapService
}

// NewMapBinding creates a new map binding
func NewMapBinding(service *services.MapService) *MapBinding {
	return &MapBinding{service: service}
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
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
	Items []MapResponse `json:"items"`
}

// MapResponse represents a map in API responses
type MapResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Problem     string   `json:"problem"`
	ProblemType string   `json:"problemType"`
	Target      string   `json:"target"`
	KeyPoints   []string `json:"keyPoints"`
	Constraints []string `json:"constraints"`
	Conclusion  string   `json:"conclusion"`
	Status      string   `json:"status"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
}

// CreateMap creates a new thinking map
func (b *MapBinding) CreateMap(ctx context.Context, req CreateMapRequest) (MapResponse, error) {
	mapModel, err := b.service.CreateMap(ctx, services.CreateMapRequest{
		Title:       req.Title,
		Problem:     req.Problem,
		ProblemType: req.ProblemType,
		Target:      req.Target,
		KeyPoints:   req.KeyPoints,
		Constraints: req.Constraints,
	})
	if err != nil {
		return MapResponse{}, err
	}
	return toMapResponse(mapModel), nil
}

// ListMaps retrieves a list of thinking maps
func (b *MapBinding) ListMaps(ctx context.Context, query MapListQuery) (MapListResponse, error) {
	result, err := b.service.ListMaps(ctx, services.MapListQuery{
		Page:        query.Page,
		Limit:       query.Limit,
		Status:      query.Status,
		ProblemType: query.ProblemType,
		Search:      query.Search,
		DateRange:   query.DateRange,
	})
	if err != nil {
		return MapListResponse{}, err
	}

	items := make([]MapResponse, len(result.Items))
	for i, m := range result.Items {
		items[i] = toMapResponse(m)
	}

	return MapListResponse{
		Total: result.Total,
		Page:  result.Page,
		Limit: result.Limit,
		Items: items,
	}, nil
}

// GetMap retrieves a specific thinking map
func (b *MapBinding) GetMap(ctx context.Context, mapID string) (MapResponse, error) {
	mapModel, err := b.service.GetMap(ctx, mapID)
	if err != nil {
		return MapResponse{}, err
	}
	return toMapResponse(mapModel), nil
}

// UpdateMap updates a thinking map
func (b *MapBinding) UpdateMap(ctx context.Context, mapID string, req UpdateMapRequest) (MapResponse, error) {
	mapModel, err := b.service.UpdateMap(ctx, mapID, services.UpdateMapRequest{
		Status:      req.Status,
		Problem:     req.Problem,
		ProblemType: req.ProblemType,
		Target:      req.Target,
		KeyPoints:   req.KeyPoints,
		Constraints: req.Constraints,
		Conclusion:  req.Conclusion,
	})
	if err != nil {
		return MapResponse{}, err
	}
	return toMapResponse(mapModel), nil
}

// DeleteMap deletes a thinking map
func (b *MapBinding) DeleteMap(ctx context.Context, mapID string) error {
	return b.service.DeleteMap(ctx, mapID)
}

// InitMapBinding initializes a MapBinding with its service dependency
// This is called by AppContext during application startup
func InitMapBinding(b *MapBinding, service *services.MapService) {
	b.service = service
}

func toMapResponse(m *models.ThinkingMap) MapResponse {
	return MapResponse{
		ID:          m.ID,
		Title:       m.Title,
		Problem:     m.Problem,
		ProblemType: m.ProblemType,
		Target:      m.Target,
		KeyPoints:   m.KeyPoints,
		Constraints: m.Constraints,
		Conclusion:  m.Conclusion,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Unix(),
		UpdatedAt:   m.UpdatedAt.Unix(),
	}
}
