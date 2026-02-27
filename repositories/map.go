/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/repositories/map.go
 */
package repositories

import (
	"context"
	"time"

	"github.com/PGshen/thinking-map-desktop/models"
	"gorm.io/gorm"
)

const whereID = "id = ?"

// ThinkingMapRepository 思维导图仓储
type ThinkingMapRepository interface {
	Create(ctx context.Context, map_ *models.ThinkingMap, rootNode *models.ThinkingNode) error
	Update(ctx context.Context, mapID string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.ThinkingMap, error)
	List(ctx context.Context, status string, problemType, search string, startTime, endTime time.Time, page, limit int) ([]*models.ThinkingMap, int64, error)
}

type thinkingMapRepository struct {
	db *gorm.DB
}

// NewThinkingMapRepository creates a new map repository
func NewThinkingMapRepository(db *gorm.DB) ThinkingMapRepository {
	return &thinkingMapRepository{db: db}
}

// Create creates a new thinking map and its root node
func (r *thinkingMapRepository) Create(ctx context.Context, thinkingMap *models.ThinkingMap, rootNode *models.ThinkingNode) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Create(thinkingMap).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Create(rootNode).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// List retrieves a list of thinking maps with pagination (no userID filter for desktop)
func (r *thinkingMapRepository) List(ctx context.Context, status string, problemType, search string, startTime, endTime time.Time, page, limit int) ([]*models.ThinkingMap, int64, error) {
	var maps []*models.ThinkingMap
	var total int64

	dbQuery := r.db.Model(&models.ThinkingMap{})
	if status != "" {
		dbQuery = dbQuery.Where("status = ?", status)
	}
	if problemType != "" {
		dbQuery = dbQuery.Where("problem_type = ?", problemType)
	}
	if search != "" {
		like := "%" + search + "%"
		dbQuery = dbQuery.Where("problem LIKE ? OR target LIKE ?", like, like)
	}

	if !startTime.IsZero() && !endTime.IsZero() {
		dbQuery = dbQuery.Where("created_at BETWEEN ? AND ?", startTime, endTime)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := dbQuery.Offset(offset).Limit(limit).Order("created_at DESC").Find(&maps).Error; err != nil {
		return nil, 0, err
	}

	return maps, total, nil
}

// FindByID retrieves a specific thinking map
func (r *thinkingMapRepository) FindByID(ctx context.Context, mapID string) (*models.ThinkingMap, error) {
	var thinkingMap models.ThinkingMap
	if err := r.db.Where(whereID, mapID).First(&thinkingMap).Error; err != nil {
		return nil, err
	}
	return &thinkingMap, nil
}

// Update updates a thinking map
func (r *thinkingMapRepository) Update(ctx context.Context, mapID string, updates map[string]interface{}) error {
	return r.db.Model(&models.ThinkingMap{}).
		Where("id = ?", mapID).
		Updates(updates).Error
}

// Delete deletes a thinking map and all its nodes
func (r *thinkingMapRepository) Delete(ctx context.Context, mapID string) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("map_id = ?", mapID).Delete(&models.ThinkingNode{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("id = ?", mapID).Delete(&models.ThinkingMap{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
