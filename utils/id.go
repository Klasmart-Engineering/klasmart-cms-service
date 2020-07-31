package utils

import "github.com/google/uuid"

func NewId()string{
	return uuid.New().String()
}
