package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

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
			values = append(values, fmt.Sprintf("%v", v))
		}
	default:
		if value := fmt.Sprint(value); value != "" {
			values = []string{fmt.Sprintf("%v", value)}
		}
	}
	return
}

func ToString(value interface{}) string {
	if v, ok := value.([]string); ok && len(v) > 0 {
		return v[0]
	} else if v, ok := value.(string); ok {
		return v
	} else if v, ok := value.([]interface{}); ok && len(v) > 0 {
		return fmt.Sprintf("%v", v[0])
	} else {
		return fmt.Sprintf("%v", value)
	}
}

func ToInt(value interface{}) int64 {
	var result string
	if v, ok := value.([]string); ok && len(v) > 0 {
		result = v[0]
	} else if v, ok := value.(string); ok {
		result = v
	} else {
		return ToInt(fmt.Sprintf("%v", value))
	}

	if i, err := strconv.ParseInt(result, 10, 64); err == nil {
		return i
	} else if result == "" {
		return 0
	} else {
		panic("failed to parse int: " + result)
	}
}

func ToUint(value interface{}) uint64 {
	var result string
	if v, ok := value.([]string); ok && len(v) > 0 {
		result = v[0]
	} else if v, ok := value.(string); ok {
		result = v
	} else {
		return ToUint(fmt.Sprintf("%v", value))
	}

	if i, err := strconv.ParseUint(result, 10, 64); err == nil {
		return i
	} else if result == "" {
		return 0
	} else {
		panic("failed to parse uint: " + result)
	}
}

func ToFloat(value interface{}) float64 {
	var result string
	if v, ok := value.([]string); ok && len(v) > 0 {
		result = v[0]
	} else if v, ok := value.(string); ok {
		result = v
	} else {
		return ToFloat(fmt.Sprintf("%v", value))
	}

	if i, err := strconv.ParseFloat(result, 64); err == nil {
		return i
	} else if result == "" {
		return 0
	} else {
		panic("failed to parse float: " + result)
	}
}
