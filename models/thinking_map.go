/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/app/models/thinking_map.go
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

// ThinkingMap 思维导图模型
type ThinkingMap struct {
	SerialID    int64          `gorm:"primaryKey;autoIncrement;column:serial_id" json:"-"`
	ID          string         `gorm:"type:varchar(36);uniqueIndex" json:"id"`
	Title       string         `json:"title" gorm:"type:varchar(255);not null"`
	Problem     string         `json:"problem" gorm:"type:text;not null"`
	ProblemType string         `json:"problem_type" gorm:"type:varchar(50)"`
	Target      string         `json:"target" gorm:"type:text"`
	KeyPoints   KeyPoints      `json:"key_points" gorm:"type:text"`
	Constraints Constraints    `json:"constraints" gorm:"type:text"`
	Conclusion  string         `json:"conclusion" gorm:"type:text"`
	Status      string         `json:"status" gorm:"type:varchar(16);not null;default:'initial'"` // initial, running, completed, deleted
	Metadata    Metadata       `json:"metadata" gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time      `gorm:"type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *ThinkingMap) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" || t.ID == uuid.Nil.String() {
		t.ID = uuid.NewString()
	}
	return nil
}

// TableName 定义表名
func (ThinkingMap) TableName() string {
	return "thinking_maps"
}

// KeyPoints 关键点列表
type KeyPoints []string

// Scan implements sql.Scanner
func (k *KeyPoints) Scan(value interface{}) error {
	if value == nil {
		*k = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		// Try string type (DuckDB may return string)
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("failed to unmarshal JSON value: %v", value)
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, k)
}

// Value implements driver.Valuer
func (k KeyPoints) Value() (driver.Value, error) {
	if len(k) == 0 {
		return "[]", nil
	}
	return json.Marshal(k)
}

// Constraints 约束列表
type Constraints []string

// Scan implements sql.Scanner
func (c *Constraints) Scan(value interface{}) error {
	if value == nil {
		*c = nil
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

// Value implements driver.Valuer
func (c Constraints) Value() (driver.Value, error) {
	if len(c) == 0 {
		return "[]", nil
	}
	return json.Marshal(c)
}

// Metadata 元数据
type Metadata map[string]interface{}

// Scan implements sql.Scanner
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
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

// Value implements driver.Valuer
func (m Metadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	return json.Marshal(m)
}
