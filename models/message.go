/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/models/message.go
 */
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Message 消息模型
type Message struct {
	SerialID       int64           `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID             string          `gorm:"type:varchar(36);uniqueIndex"`
	ParentID       string          `gorm:"type:varchar(36);index"`
	ConversationID string          `gorm:"type:varchar(36);index"`
	MessageType    MsgType         `gorm:"type:varchar(20);not null;default:text"` // text, rag, notice, action
	Role           schema.RoleType `gorm:"type:varchar(48)"`
	Content        MessageContent  `gorm:"type:text;not null"`
	Metadata       Metadata        `gorm:"type:text"`
	CreatedAt      time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time       `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt      gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" || m.ID == uuid.Nil.String() {
		m.ID = uuid.NewString()
	}
	return nil
}

func (Message) TableName() string {
	return "messages"
}

// 消息类型
type MsgType string

const (
	MsgTypeText    MsgType = "text"
	MsgTypeRAG     MsgType = "rag"
	MsgTypeNotice  MsgType = "notice"
	MsgTypeAction  MsgType = "action"
	MsgTypeThought MsgType = "thought"
	MsgTypePlan    MsgType = "plan"
)

// NoticeType 通知类型
type NoticeType string

const (
	NoticeTypeError   NoticeType = "error"
	NoticeTypeWarning NoticeType = "warning"
	NoticeTypeSuccess NoticeType = "success"
	NoticeTypeInfo    NoticeType = "info"
)

// Notice 通知信息
type Notice struct {
	Type    NoticeType `json:"type"`
	Name    string     `json:"name"`
	Content string     `json:"content"`
}

type Action struct {
	Name   string         `json:"name"`
	URL    string         `json:"url"`
	Method string         `json:"method"`
	Param  map[string]any `json:"param,omitempty"`
}

type Plan struct {
	Steps []PlanStep `json:"steps"`
}

// PlanStep 计划步骤
type PlanStep struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	AssignedSpecialist string `json:"assignedSpecialist"`
	Status             string `json:"status"`
}

// MessageContent 消息内容
type MessageContent struct {
	Text    string   `json:"text,omitempty"`
	Thought string   `json:"thought,omitempty"`
	RagID   string   `json:"rag,omitempty"` // 外键
	Notice  *Notice  `json:"notice,omitempty"`
	Action  []Action `json:"action,omitempty"`
	Plan    *Plan    `json:"plan,omitempty"`
}

// MessageContent implements Scanner interface
func (m *MessageContent) Scan(value interface{}) error {
	if value == nil {
		*m = MessageContent{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshal JSON value: %v", value)
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, m)
}

// MessageContent implements Valuer interface
func (m MessageContent) Value() (driver.Value, error) {
	return json.Marshal(m)
}
