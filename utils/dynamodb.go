package utils

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"reflect"
)

func GetUpdateBuilder(param interface{}) (expression.UpdateBuilder, error) {
	var result = expression.UpdateBuilder{}
	t := reflect.TypeOf(param)
	val := reflect.ValueOf(param)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		return result, errors.New("Check type error not Struct")
	}
	fieldNum := t.NumField()

	for i := 0; i < fieldNum; i++ {
		fieldName := t.Field(i).Name
		fieldVal := val.FieldByName(fieldName).Interface()

		dynamoFlag, ok := t.Field(i).Tag.Lookup("dynamodbav")
		if !ok {
			continue
		}
		result = result.Set(expression.Name(dynamoFlag), expression.Value(fieldVal))
	}
	return result, nil
}
