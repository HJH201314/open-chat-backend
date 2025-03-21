package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type MilliTime struct {
	time.Time
}

// MarshalJSON 实现 JSON 序列化
func (t *MilliTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.UnixMilli())
}

// UnmarshalJSON 实现 JSON 反序列化
func (t *MilliTime) UnmarshalJSON(data []byte) error {
	var ms int64
	if err := json.Unmarshal(data, &ms); err != nil {
		return err
	}
	t.Time = time.UnixMilli(ms)
	return nil
}

// Scan 实现数据库读取
func (t *MilliTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		t.Time = v
	case []byte:
		return parseTimeString(t, string(v))
	case string:
		return parseTimeString(t, v)
	case nil:
		t.Time = time.Time{}
	default:
		return fmt.Errorf("无法转换类型 %T 到 MilliTime", value)
	}
	return nil
}

// 辅助函数：解析时间字符串
func parseTimeString(t *MilliTime, s string) error {
	layouts := []string{
		time.RFC3339,
		"2025-01-02 03:04:05.666666666",
		"2025-01-02T03:04:05.666666666Z07:00",
	}

	var err error
	for _, layout := range layouts {
		t.Time, err = time.ParseInLocation(layout, s, time.UTC)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("无法解析时间字符串: %s", s)
}

// Value 实现数据库存储
func (t *MilliTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time.UTC(), nil
}
