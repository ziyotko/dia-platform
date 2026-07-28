package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime is a wrapper for time.Time for JSON serialization
type LocalTime struct {
	time.Time
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	parsed, err := time.ParseInLocation(`"2006-01-02 15:04:05"`, string(data), time.Local)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func (t LocalTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *LocalTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case []byte:
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", string(v), time.Local)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	case string:
		parsed, err := time.ParseInLocation("2006-01-02 15:04:05", v, time.Local)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	}
	return fmt.Errorf("cannot scan type %T into LocalTime", value)
}
