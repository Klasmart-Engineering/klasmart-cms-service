package utils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewID()string{
	return primitive.NewObjectID().String()
}
