package utils

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewID() string {
	return primitive.NewObjectID().Hex()
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
