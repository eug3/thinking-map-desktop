/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/repositories/message.go
 */
package repositories

import (
	"context"

	"github.com/PGshen/thinking-map-desktop/models"
	"gorm.io/gorm"
)

// MessageRepository 消息仓储接口
type MessageRepository interface {
	Create(ctx context.Context, message *models.Message) error
	CreateInTx(ctx context.Context, tx *gorm.DB, message *models.Message) error
	Update(ctx context.Context, message *models.Message) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.Message, error)
	ByParentID(ctx context.Context, parentID string) ([]*models.Message, error)
	ByConversationID(ctx context.Context, conversationID string) ([]*models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) CreateInTx(ctx context.Context, tx *gorm.DB, message *models.Message) error {
	return tx.WithContext(ctx).Create(message).Error
}

func (r *messageRepository) Update(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *messageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where(whereID, id).Delete(&models.Message{}).Error
}

func (r *messageRepository) FindByID(ctx context.Context, id string) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).Where(whereID, id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *messageRepository) ByParentID(ctx context.Context, parentID string) ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.WithContext(ctx).Where("parent_id = ?", parentID).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepository) ByConversationID(ctx context.Context, conversationID string) ([]*models.Message, error) {
	var messages []*models.Message
	err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
