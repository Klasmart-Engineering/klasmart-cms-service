package dynamodbhelper

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

type Condition struct {
	Pager Pager

	PrimaryKey KeyValue
	SortKey    KeyValue

	CompareType CompareType
	IndexName   string
}

func (s Condition) KeyBuilder() expression.KeyConditionBuilder {
	var (
		primaryKey expression.KeyConditionBuilder
		sortKey    expression.KeyConditionBuilder
	)
	primaryKey = expression.Key(s.PrimaryKey.Key).Equal(expression.Value(s.PrimaryKey.Value))

	switch s.CompareType {
	case SortKeyEqual:
		sortKey = expression.Key(s.SortKey.Key).Equal(expression.Value(s.SortKey.Value))
	case SortKeyGreaterThanEqual:
		sortKey = expression.Key(s.SortKey.Key).GreaterThanEqual(expression.Value(s.SortKey.Value))
	}

	return primaryKey.And(sortKey)
}

type KeyValue struct {
	Key   string
	Value interface{}
}

type CompareType string

const (
	SortKeyNone             CompareType = "None"
	SortKeyEqual            CompareType = "Equal"
	SortKeyGreaterThanEqual CompareType = "GreaterThanEqual"
)

type Pager struct {
	PageSize int64
	LastKey  string
}

const (
	DefaultPageSize = 10
)
