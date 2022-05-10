package external

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/KL-Engineering/common-log/log"
	"github.com/KL-Engineering/kidsloop-cms-service/constant"
)

const (
	omitempty string = "omitempty"
	noquoted  string = "noquoted"
)

func marshalFilter(v interface{}) (string, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	//case reflect.Pointer:
	case reflect.Struct:
		result, err := structEncode(rv)
		if err != nil {
			return "", err
		}
		return result, nil
	//case reflect.Slice:
	//case reflect.String:
	//case reflect.Bool:
	//case reflect.Int:
	default:
		return "", constant.ErrGraphQLUnsupported
	}
}

func structEncode(v reflect.Value) (string, error) {
	var names []string
	var values []string
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		var opts string
		name := v.Type().Field(i).Name
		f := v.Field(i)
		tags, ok := t.Field(i).Tag.Lookup("gqls")
		if ok {
			name, opts, _ = strings.Cut(tags, ",")
		}
		switch f.Kind() {
		case reflect.Pointer:
			if f.IsNil() {
				continue
			}
			result, err := pointerEncode(f.Elem())
			if err != nil {
				return "", err
			}
			if strings.Contains(opts, omitempty) && result == "" {
				continue
			}
			names = append(names, name)
			values = append(values, result)
		case reflect.Struct:
			result, err := structEncode(f)
			if err != nil {
				return "", err
			}
			if strings.Contains(opts, omitempty) && result == "" {
				continue
			}
			names = append(names, name)
			values = append(values, result)
		case reflect.Slice:
			if f.IsNil() {
				continue
			}
			result, err := sliceEncode(f)
			if err != nil {
				return "", err
			}
			if strings.Contains(opts, omitempty) && result == "" {
				continue
			}
			names = append(names, name)
			values = append(values, result)
		case reflect.String:
			value := f.String()
			if strings.Contains(opts, omitempty) && value == "" {
				continue
			}
			names = append(names, name)
			if !strings.Contains(opts, noquoted) {
				value = "\"" + value + "\""
			}
			values = append(values, value)
		case reflect.Bool:
			value := f.Bool()
			if strings.Contains(opts, omitempty) && !value {
				continue
			}
			names = append(names, name)
			if value {
				values = append(values, "true")
			} else {
				values = append(values, "false")
			}
		case reflect.Int:
			value := f.Int()
			if strings.Contains(opts, omitempty) && value == 0 {
				continue
			}
			names = append(names, name)
			values = append(values, strconv.FormatInt(value, 10))
		}
	}
	if len(names) == 0 || len(values) == 0 {
		return "", nil
	}
	if len(names) != len(values) {
		return "", constant.ErrGraphQLUnMatch
	}

	namesValues := make([]string, 0, len(names))
	for i := 0; i < len(names); i++ {
		nameValue := fmt.Sprintf("%s:%s", names[i], values[i])
		namesValues = append(namesValues, nameValue)
	}
	return fmt.Sprintf("{%s}", strings.Join(namesValues, ",")), nil
}

func pointerEncode(p reflect.Value) (string, error) {
	switch p.Kind() {
	case reflect.Pointer:
		if p.IsNil() {
			return "", nil
		}
		result, err := pointerEncode(p.Elem())
		if err != nil {
			return "", err
		}
		return result, nil
	case reflect.Struct:
		result, err := structEncode(p)
		if err != nil {
			return "", err
		}
		return result, nil
	case reflect.Slice:
		if p.IsNil() {
			return "", nil
		}
		result, err := sliceEncode(p)
		if err != nil {
			return "", err
		}
		return result, nil
	case reflect.String:
		value := p.String()
		value = "\"" + value + "\""
		return value, nil
	case reflect.Bool:
		value := p.Bool()
		if value {
			return "true", nil
		}
		return "false", nil
	case reflect.Int:
		value := p.Int()
		return strconv.FormatInt(value, 10), nil
	default:
		return "", constant.ErrGraphQLUnsupported
	}
}

func sliceEncode(s reflect.Value) (string, error) {
	var values []string
	for i := 0; i < s.Len(); i++ {
		e := s.Index(i)
		switch e.Kind() {
		case reflect.Pointer:
			if e.IsNil() {
				continue
			}
			result, err := pointerEncode(e.Elem())
			if err != nil {
				return "", err
			}
			if result == "" {
				continue
			}
			values = append(values, result)
		case reflect.Struct:
			result, err := structEncode(e)
			if err != nil {
				return "", err
			}
			if result == "" {
				continue
			}
			values = append(values, result)
		case reflect.Slice:
			if e.IsNil() {
				continue
			}
			result, err := sliceEncode(e)
			if err != nil {
				return "", err
			}
			if result == "" {
				continue
			}
			values = append(values, result)
		case reflect.String:
			value := e.String()
			if value == "" {
				continue
			}
			values = append(values, value)
		case reflect.Bool:
			value := e.Bool()
			if value {
				values = append(values, "true")
			} else {
				values = append(values, "false")
			}
		case reflect.Int:
			value := e.Int()
			values = append(values, strconv.FormatInt(value, 10))
		}
	}
	if len(values) == 0 {
		return "", nil
	}

	return fmt.Sprintf("[%s]", strings.Join(values, ",")), nil
}

func marshalFiled(ctx context.Context, v interface{}) (string, error) {
	rv := reflect.ValueOf(v)
	var values []string
	if rv.Kind() == reflect.Pointer {
		rev := rv.Elem()
		if rev.Kind() == reflect.Slice {
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
