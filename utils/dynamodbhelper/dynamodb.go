package dynamodbhelper

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
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
	Pager utils.Pager

	PrimaryKey KeyValue
	SortKey    KeyValue
	IndexName  string
}

func (s Condition) GetKeyConditionBuilder(buildType string) expression.KeyConditionBuilder {
	switch buildType {
	case BuilderTypeKeyEqule:
		primaryKey := expression.Key(s.PrimaryKey.Key).Equal(expression.Value(s.PrimaryKey.Value))
		sortKey := expression.Key(s.SortKey.Key).Equal(expression.Value(s.SortKey.Value))
		return expression.KeyAnd(primaryKey, sortKey)
	default:
		primaryKey := expression.Key(s.PrimaryKey.Key).Equal(expression.Value(s.PrimaryKey.Value))
		return primaryKey
	}
}

type KeyValue struct {
	Key   string
	Value interface{}
}

const (
	BuilderTypeKeyEqule = "KeyEqule"
)
