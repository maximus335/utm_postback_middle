package types

import (
	"database/sql/driver"
	"encoding/json"
)

type JsonbType map[string]interface{}

// Value implements the driver Valuer interface.
func (a JsonbType) Value() (driver.Value, error) {
    return json.Marshal(a)
}

// Scan implements the Scanner interface.
func (a *JsonbType) Scan(value interface{}) error {
    b, ok := value.([]byte)
    if !ok {
        return nil
    }

    return json.Unmarshal(b, &a)
}
