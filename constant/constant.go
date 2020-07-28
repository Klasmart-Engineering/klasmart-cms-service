package constant

import "errors"

const (
	TableNameTag = "tags"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicateRecord = errors.New("duplicate record")
	ErrExceededLimit = errors.New("exceeded limit")
)
