package external

import (
	"context"
	"fmt"
	"gitlab.badanamu.com.cn/calmisland/common-log/log"
	"reflect"
	"strings"
)

const (
	omitempty string = "omitempty"
	noquoted  string = "noquoted"
)

func marshalFiled(ctx context.Context, v interface{}) (string, error) {
	rv := reflect.ValueOf(v)
	var values []string
	if rv.Kind() == reflect.Pointer {
		rev := rv.Elem()
		if rev.Kind() == reflect.Slice {
			result, err := sliceField(ctx, rev)
			if err != nil {
				// TODO
			}
			values = append(values, result)
			//for i := 0; i < rev.Type().Elem().NumField(); i++ {
			//	f := rev.Type().Elem().Field(i)
			//	name := f.Name
			//	tags, ok := f.Tag.Lookup("gqls")
			//	if ok {
			//		name, _, _ = strings.Cut(tags, ",")
			//	}
			//	//fmt.Println(name, f.Type.Kind())
			//	values = append(values, name)
			//	switch f.Type.Kind() {
			//	case reflect.Slice:
			//		result, err := structField(ctx, f.Type.Elem())
			//		if err != nil {
			//			log.Error(ctx, "fields: struct failed",
			//				log.Err(err),
			//				log.String("slice", name))
			//			return "", err
			//		}
			//		values = append(values, result)
			//	case reflect.Struct:
			//		result, err := structField(ctx, f.Type)
			//		if err != nil {
			//			log.Error(ctx, "fields: struct failed",
			//				log.Err(err),
			//				log.String("struct", name))
			//			return "", err
			//		}
			//		values = append(values, result)
			//	}
			//}
		}
	} else if rv.Kind() == reflect.Map {
		rev := rv.Type().Elem()
		if rev.Kind() == reflect.Slice {
			elemType := rev.Elem()
			if elemType.Kind() == reflect.Struct {
				result, err := structField(ctx, elemType)
				if err != nil {
					// TODO
				}
				values = append(values, result)
			}
			fmt.Println(rev)
		}

	}
	return strings.Join(values, "\n"), nil
}

func sliceField(ctx context.Context, rev reflect.Value) (string, error) {
	var values []string
	for i := 0; i < rev.Type().Elem().NumField(); i++ {
		f := rev.Type().Elem().Field(i)
		name := f.Name
		tags, ok := f.Tag.Lookup("gqls")
		if ok {
			name, _, _ = strings.Cut(tags, ",")
		}
		//fmt.Println(name, f.Type.Kind())
		values = append(values, name)
		switch f.Type.Kind() {
		case reflect.Slice:
			result, err := structField(ctx, f.Type.Elem())
			if err != nil {
				log.Error(ctx, "fields: struct failed",
					log.Err(err),
					log.String("slice", name))
				return "", err
			}
			values = append(values, result)
		case reflect.Struct:
			result, err := structField(ctx, f.Type)
			if err != nil {
				log.Error(ctx, "fields: struct failed",
					log.Err(err),
					log.String("struct", name))
				return "", err
			}
			values = append(values, result)
		}
	}
	return strings.Join(values, "\n"), nil
}

func structField(ctx context.Context, t reflect.Type) (string, error) {
	names := []string{"{"}
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		name := ft.Name
		tags, ok := ft.Tag.Lookup("gqls")
		if ok {
			name, _, _ = strings.Cut(tags, ",")
		}
		names = append(names, name)
		switch ft.Type.Kind() {
		case reflect.Struct:
			result, err := structField(ctx, ft.Type)
			if err != nil {
				log.Error(ctx, "struct failed",
					log.Err(err),
					log.String("struct", name))
				return "", err
			}
			names = append(names, result)
		case reflect.Slice:
			result, err := structField(ctx, ft.Type.Elem())
			if err != nil {
				log.Error(ctx, "struct failed",
					log.Err(err),
					log.String("slice", name))
				return "", err
			}
			names = append(names, result)
		default:
		}
	}
	names = append(names, "}")
	return strings.Join(names, "\n"), nil
}
