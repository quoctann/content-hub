package http

import (
	"encoding/json"

	"github.com/quoctann/content-hub/internal/domain"
)

// NullableString is a PATCH-style JSON field with three states:
//
//	field omitted        -> Set = false (leave the column unchanged)
//	"field": null        -> Set = true,  Value = nil (clear the column)
//	"field": "some text" -> Set = true,  Value = &"some text"
//
// A plain *string cannot tell the first two apart, which is why clearing a
// field from the admin editor used to be silently ignored.
type NullableString struct {
	Set   bool
	Value *string
}

// UnmarshalJSON is only called when the key is present (including for null),
// so reaching it at all means the client sent the field.
func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = &s
	return nil
}

func (n NullableString) toDomain() domain.Optional[*string] {
	return domain.Optional[*string]{Set: n.Set, Value: n.Value}
}
