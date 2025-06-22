package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MatchMode int
type SearchFieldType int
type SortOrder int

func enumValue(b []byte, names map[string]int) (int, error) {
	var text string
	if json.Unmarshal(b, &text) == nil {
		if n, ok := names[strings.ToLower(text)]; ok {
			return n, nil
		}
		return 0, fmt.Errorf("%w: unknown enum value", ErrInvalid)
	}
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return 0, ErrInvalid
	}
	for _, allowed := range names {
		if n == allowed {
			return n, nil
		}
	}
	return 0, ErrInvalid
}
func (v *MatchMode) UnmarshalJSON(b []byte) error {
	n, err := enumValue(b, map[string]int{"equals": 1, "startswith": 2, "contains": 3})
	*v = MatchMode(n)
	return err
}
func (v *SearchFieldType) UnmarshalJSON(b []byte) error {
	n, err := enumValue(b, map[string]int{"string": 0, "number": 1, "date": 2, "boolean": 3, "array": 4})
	*v = SearchFieldType(n)
	return err
}
func (v *SortOrder) UnmarshalJSON(b []byte) error {
	n, err := enumValue(b, map[string]int{"asc": 1, "desc": -1})
	*v = SortOrder(n)
	return err
}
