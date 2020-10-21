package entity

import "strings"

type NullInts struct {
	Ints  []int
	Valid bool
}

type NullStrings struct {
	Strings []string
	Valid   bool
}

func (s NullStrings) SQLPlaceHolder() string {
	if len(s.Strings) == 0 && s.Valid {
		return "null"
	}

	return strings.TrimSuffix(strings.Repeat("?,", len(s.Strings)), ",")
}
func (s NullStrings) ToInterfaceSlice() []interface{} {
	slice := make([]interface{}, len(s.Strings))
	for index, value := range s.Strings {
		slice[index] = value
	}

	return slice
}
func SplitStringToNullStrings(str string) NullStrings {
	if strings.TrimSpace(str) != "" {
		strArr := strings.Split(str, ",")
		return NullStrings{Strings: strArr, Valid: len(strArr) > 0}
	}
	return NullStrings{}
}

type NullString struct {
	String string
	Valid  bool
}

type NullInt struct {
	Int   int
	Valid bool
}
type NullInt64 struct {
	Int64 int64
	Valid bool
}
type IDResponse struct {
	ID string `json:"id"`
}
