package utils

import (
	"database/sql/driver"
	"encoding/json"
)

type SQLJSONStringArray []string

func (a *SQLJSONStringArray) Value() (driver.Value, error) {
	if a == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(*a)
}

func (a *SQLJSONStringArray) Scan(src interface{}) error {
	if v, ok := src.([]byte); ok && string(v) != "" {
		return json.Unmarshal(v, a)
	}
	return nil
}
