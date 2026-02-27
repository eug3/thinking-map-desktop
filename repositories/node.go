/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/repositories/node.go
 */
package repositories

import (
	"context"
	"errors"

	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ThinkingNodeRepository 节点仓储接口
type ThinkingNodeRepository interface {
	Create(ctx context.Context, node *models.ThinkingNode) error
	Update(ctx context.Context, node *models.ThinkingNode) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.ThinkingNode, error)
	ByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*models.ThinkingNode, error)
	FindByIDs(ctx context.Context, ids []string) ([]*models.ThinkingNode, error)
	ByMapID(ctx context.Context, mapID string) ([]*models.ThinkingNode, error)
	ByParentID(ctx context.Context, parentID string) ([]*models.ThinkingNode, error)
	UpdatePosition(ctx context.Context, id string, position models.Position) error
	DeleteByParentID(ctx context.Context, parentID string) error
}

type thinkingNodeRepository struct {
	db *gorm.DB
}

// NewThinkingNodeRepository creates a new node repository
func NewThinkingNodeRepository(db *gorm.DB) ThinkingNodeRepository {
	return &thinkingNodeRepository{db: db}
}

func (r *thinkingNodeRepository) Create(ctx context.Context, node *models.ThinkingNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

func (r *thinkingNodeRepository) Update(ctx context.Context, node *models.ThinkingNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

func (r *thinkingNodeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where(whereID, id).Delete(&models.ThinkingNode{}).Error
}

func (r *thinkingNodeRepository) FindByID(ctx context.Context, id string) (*models.ThinkingNode, error) {
	var node models.ThinkingNode
	if id == uuid.Nil.String() {
		err := errors.New("record not found")
		return &node, err
	}
	err := r.db.WithContext(ctx).Where(whereID, id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ByIDForUpdate uses row-level lock to find node
func (r *thinkingNodeRepository) ByIDForUpdate(ctx context.Context, tx *gorm.DB, id string) (*models.ThinkingNode, error) {
	var node models.ThinkingNode
	err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where(whereID, id).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *thinkingNodeRepository) DeleteByParentID(ctx context.Context, parentID string) error {
	return r.db.WithContext(ctx).Where("parent_id = ?", parentID).Delete(&models.ThinkingNode{}).Error
}

// FindByIDs retrieves multiple ThinkingNode records by their IDs
func (r *thinkingNodeRepository) FindByIDs(ctx context.Context, ids []string) ([]*models.ThinkingNode, error) {
	var nodes []*models.ThinkingNode
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, err
}

// ByMapID retrieves all nodes for a map
func (r *thinkingNodeRepository) ByMapID(ctx context.Context, mapID string) ([]*models.ThinkingNode, error) {
	var nodes []*models.ThinkingNode
	err := r.db.WithContext(ctx).Where("map_id = ?", mapID).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// ByParentID retrieves all child nodes of a parent
func (r *thinkingNodeRepository) ByParentID(ctx context.Context, parentID string) ([]*models.ThinkingNode, error) {
	var nodes []*models.ThinkingNode
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&nodes).Error
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// UpdatePosition updates node position
func (r *thinkingNodeRepository) UpdatePosition(ctx context.Context, id string, position models.Position) error {
	return r.db.WithContext(ctx).Model(&models.ThinkingNode{}).
		Where("id = ?", id).
		Update("position", position).Error
}
