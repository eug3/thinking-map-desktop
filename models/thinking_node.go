/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/models/thinking_node.go
 */
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ThinkingNode 思维节点模型
type ThinkingNode struct {
	SerialID      int64            `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID            string           `gorm:"type:varchar(36);uniqueIndex"`
	MapID         string           `gorm:"type:varchar(36);not null;index"`
	ParentID      string           `gorm:"type:varchar(36);index"`
	NodeType      string           `gorm:"type:varchar(50);not null"` // root, analysis, conclusion, custom
	Question      string           `gorm:"type:text;not null"`
	Target        string           `gorm:"type:text"`
	Context       DependentContext `gorm:"type:text"`
	Decomposition Decomposition    `gorm:"type:text"`
	Conclusion    Conclusion       `gorm:"type:text"`
	Status        string           `gorm:"type:varchar(16);default:'initial'"` // initial, pending, running, completed, error
	Position      Position         `gorm:"type:text"`
	Metadata      Metadata         `gorm:"type:text"`
	Dependencies  Dependencies     `gorm:"type:text"`
	CreatedAt     time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time        `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (t *ThinkingNode) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" || t.ID == uuid.Nil.String() {
		t.ID = uuid.NewString()
	}
	return nil
}

func (ThinkingNode) TableName() string {
	return "thinking_nodes"
}

// Position 节点位置信息
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Dependency 节点依赖信息
type Dependencies []string

// DependentContext 依赖上下文
type DependentContext struct {
	Ancestor    []NodeContext `json:"ancestor,omitempty"`    // 祖先节点
	PrevSibling []NodeContext `json:"prevSibling,omitempty"` // 前置兄弟节点
	Children    []NodeContext `json:"children,omitempty"`    // 子节点
}

// NodeContext 节点上下文
type NodeContext struct {
	Question   string `json:"question"`
	Target     string `json:"target"`
	Conclusion string `json:"conclusion"`
	Abstract   string `json:"abstract"`
	Status     string `json:"status"`
}

type Decomposition struct {
	IsDecomposed   bool   `json:"isDecomposed"`   // 是否分解
	ConversationID string `json:"conversationID"` // 对话ID
	LastMessageID  string `json:"lastMessageID"`  // 最后一条消息ID
}

// Scan implements the Scanner interface for Decomposition
func (d *Decomposition) Scan(value interface{}) error {
	if value == nil {
		*d = Decomposition{}
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
	return json.Unmarshal(bytes, d)
}

// Value implements the Valuer interface for Decomposition
func (d Decomposition) Value() (driver.Value, error) {
	return json.Marshal(d)
}

type Conclusion struct {
	ConversationID string `json:"conversationID"` // 对话ID
	LastMessageID  string `json:"lastMessageID"`  // 最后一条消息ID
	Content        string `json:"content"`        // 最终结论
}

// Scan implements the Scanner interface for Conclusion
func (c *Conclusion) Scan(value interface{}) error {
	if value == nil {
		*c = Conclusion{}
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
	return json.Unmarshal(bytes, c)
}

// Value implements the Valuer interface for Conclusion
func (c Conclusion) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Position implements Scanner interface
func (p *Position) Scan(value interface{}) error {
	if value == nil {
		*p = Position{}
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
	return json.Unmarshal(bytes, p)
}

// Position implements Valuer interface
func (p Position) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Dependencies implements Scanner interface
func (d *Dependencies) Scan(value interface{}) error {
	if value == nil {
		*d = nil
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
	return json.Unmarshal(bytes, d)
}

// Dependencies implements Valuer interface
func (d Dependencies) Value() (driver.Value, error) {
	if len(d) == 0 {
		return "[]", nil
	}
	return json.Marshal(d)
}

// Scan implements the Scanner interface for Context
func (c *DependentContext) Scan(value interface{}) error {
	if value == nil {
		*c = DependentContext{}
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
	return json.Unmarshal(bytes, c)
}

// Value implements the Valuer interface for Context
func (c DependentContext) Value() (driver.Value, error) {
	return json.Marshal(c)
}
