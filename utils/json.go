package utils

import (
	"errors"
	"reflect"
)

func JsonMerge(dst interface{}, src interface{}) error {
	dstType := reflect.TypeOf(dst)
	srcType := reflect.TypeOf(src)

	if dstType.Kind() != reflect.Ptr || srcType.Kind() != reflect.Ptr {
		return errors.New("The parameters must be of type pointer")
	}
	dstValue := reflect.ValueOf(dst)
	srcValue := reflect.ValueOf(src)

	dstType = dstType.Elem()
	dstValue = dstValue.Elem()

	srcType = srcType.Elem()
	srcValue = srcValue.Elem()

	if dstType.Kind() != reflect.Struct || srcType.Kind() != reflect.Struct {
		return errors.New("Check type error not Struct")
	}
	if dstType != srcType {
		return errors.New("Two values of different types cannot be compared")
	}
	if !srcValue.IsValid() {
		return errors.New("src is Invalid")
	}
	if !dstValue.IsValid() {
		return errors.New("dst is Invalid")
	}
	fieldNum := srcType.NumField()
	for i := 0; i < fieldNum; i++ {
		//fieldName := srcType.Field(i).Name
		fieldValue := dstValue.Field(i)
		if fieldValue.IsZero() {
			srcValue := srcValue.Field(i)
			if dstValue.Field(i).CanSet() {
				dstValue.Field(i).Set(srcValue)
			}
		}
	}
	return nil
}
