package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// RequestLog's Model， for database request_logs
type RequestLog struct {
	ID              uint64         `gorm:"column:id;type:serial;primaryKey" json:"id"`
	RequestID       string         `gorm:"column:request_id;type:varchar(64);not null;index" json:"request_id"`
	Method          string         `gorm:"column:method;type:varchar(10);not null" json:"method"`
	Path            string         `gorm:"column:path;type:varchar(255);not null;index" json:"path"`
	QueryString     string         `gorm:"column:query_string;type:text" json:"query_string"`
	StatusCode      int            `gorm:"column:status_code;not null" json:"status_code"`
	RemoteIP        string         `gorm:"column:remote_ip;type:varchar(45);not null" json:"remote_ip"`
	UserAgent       string         `gorm:"column:user_agent;type:text" json:"user_agent"`
	RequestTime     float64        `gorm:"column:request_time;not null" json:"request_time"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now();index" json:"created_at"`
	FileName        string         `gorm:"column:file_name;type:varchar(255)" json:"file_name"`
	FileSize        int64          `gorm:"column:file_size" json:"file_size"`
	FileContentJSON JSONRawMessage `gorm:"column:file_content_json;type:jsonb" json:"file_content_json"`
}

func (RequestLog) TableName() string {
	return "request_logs"
}

type JSONRawMessage json.RawMessage

// driver.Valuer interface，change JSONRawMessage into string
func (j JSONRawMessage) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

// sql.Scanner interface，change database's JSON string into JSONRawMessage
func (j *JSONRawMessage) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), j)
	}
	return json.Unmarshal(bytes, j)
}

func AutoMigrateRequestLog(db *gorm.DB) error {
	return db.AutoMigrate(&RequestLog{})
}
