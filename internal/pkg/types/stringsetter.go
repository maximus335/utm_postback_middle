package types

import (
	"encoding/json"
)

type StringSetter struct {
	Value string
	Valid bool
}

// UnmarshalJSON for StringSetter
func (s *StringSetter) UnmarshalJSON(b []byte) error {
	s.Valid = string(b) != "null"
	e := json.Unmarshal(b, &s.Value)
	return e
}
