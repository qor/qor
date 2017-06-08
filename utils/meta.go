package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

// NewValue new struct value with reflect type
func NewValue(t reflect.Type) (v reflect.Value) {
	v = reflect.New(t)
	ov := v
	for t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
		e := reflect.New(t)
		v.Set(e)
	}

	if e := v.Elem(); e.Kind() == reflect.Map && e.IsNil() {
		v.Elem().Set(reflect.MakeMap(v.Elem().Type()))
	}
	return ov
}

// ToArray get array from value, will ignore blank string to convert it to array
func ToArray(value interface{}) (values []string) {
	switch value := value.(type) {
	case []string:
		values = []string{}
		for _, v := range value {
			if v != "" {
				values = append(values, v)
			}
		}
	case []interface{}:
		for _, v := range value {
			values = append(values, fmt.Sprint(v))
		}
	default:
		if value := fmt.Sprint(value); value != "" {
			values = []string{value}
		}
	}
	return
}

// ToString get string from value, if passed value is a slice, will use the first element
func ToString(value interface{}) string {
	if v, ok := value.([]string); ok {
		for _, s := range v {
			if s != "" {
				return s
			}
		}
		return ""
	} else if v, ok := value.(string); ok {
		return v
	} else if v, ok := value.([]interface{}); ok {
		for _, s := range v {
			if fmt.Sprint(s) != "" {
				return fmt.Sprint(s)
			}
		}
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// ToInt get int from value, if passed value is empty string, result will be 0
func ToInt(value interface{}) int64 {
	if result := ToString(value); result == "" {
		return 0
	} else if i, err := strconv.ParseInt(result, 10, 64); err == nil {
		return i
	} else {
		panic("failed to parse int: " + result)
	}
}

// ToUint get uint from value, if passed value is empty string, result will be 0
func ToUint(value interface{}) uint64 {
	if result := ToString(value); result == "" {
		return 0
	} else if i, err := strconv.ParseUint(result, 10, 64); err == nil {
		return i
	} else {
		panic("failed to parse uint: " + result)
	}
}

// ToFloat get float from value, if passed value is empty string, result will be 0
func ToFloat(value interface{}) float64 {
	if result := ToString(value); result == "" {
		return 0
	} else if i, err := strconv.ParseFloat(result, 64); err == nil {
		return i
	} else {
		panic("failed to parse float: " + result)
	}
}
