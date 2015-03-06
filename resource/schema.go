package resource

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/responder"
	"github.com/qor/qor/roles"
)

func convertMapToMetaValues(values map[string]interface{}, metaors []Metaor) (*MetaValues, error) {
	metaValues := &MetaValues{}
	metaorMap := make(map[string]Metaor)
	for _, metaor := range metaors {
		metaorMap[metaor.GetName()] = metaor
	}

	for key, value := range values {
		var metaValue *MetaValue
		metaor := metaorMap[key]

		switch result := value.(type) {
		case map[string]interface{}:
			if children, err := convertMapToMetaValues(result, metaor.GetMetas()); err == nil {
				metaValue = &MetaValue{Name: key, Meta: metaor, MetaValues: children}
			}
		case []interface{}:
			for _, r := range result {
				if mr, ok := r.(map[string]interface{}); ok {
					if children, err := convertMapToMetaValues(mr, metaor.GetMetas()); err == nil {
						metaValue := &MetaValue{Name: key, Meta: metaor, MetaValues: children}
						metaValues.Values = append(metaValues.Values, metaValue)
					}
				} else {
					metaValue := &MetaValue{Name: key, Value: result, Meta: metaor}
					metaValues.Values = append(metaValues.Values, metaValue)
					break
				}
			}
		default:
			metaValue = &MetaValue{Name: key, Value: value, Meta: metaor}
		}

		if metaValue != nil {
			metaValues.Values = append(metaValues.Values, metaValue)
		}
	}
	return metaValues, nil
}

func ConvertJSONToMetaValues(reader io.Reader, metaors []Metaor) (*MetaValues, error) {
	decoder := json.NewDecoder(reader)
	values := map[string]interface{}{}
	if err := decoder.Decode(&values); err == nil {
		return convertMapToMetaValues(values, metaors)
	} else {
		return nil, err
	}
}

var (
	isCurrentLevel = regexp.MustCompile("^[^.]+$")
	isNextLevel    = regexp.MustCompile(`^(([^.\[\]]+)(\[\d+\])?)(?:\.([^.]+)+)$`)
)

func ConvertFormToMetaValues(request *http.Request, metaors []Metaor, prefix string) (*MetaValues, error) {
	metaValues := &MetaValues{}
	metaorsMap := map[string]Metaor{}
	convertedNextLevel := map[string]bool{}
	for _, metaor := range metaors {
		metaorsMap[metaor.GetName()] = metaor
	}

	newMetaValue := func(key string, value interface{}) {
		if strings.HasPrefix(key, prefix) {
			var metaValue *MetaValue
			key = strings.TrimPrefix(key, prefix)

			if matches := isCurrentLevel.FindStringSubmatch(key); len(matches) > 0 {
				name := matches[0]
				metaValue = &MetaValue{Name: name, Value: value, Meta: metaorsMap[name]}
			} else if matches := isNextLevel.FindStringSubmatch(key); len(matches) > 0 {
				name := matches[1]
				if _, ok := convertedNextLevel[name]; !ok {
					convertedNextLevel[name] = true
					metaor := metaorsMap[matches[2]]
					if children, err := ConvertFormToMetaValues(request, metaor.GetMetas(), prefix+name+"."); err == nil {
						metaValue = &MetaValue{Name: matches[2], Meta: metaor, MetaValues: children}
					}
				}
			}

			if metaValue != nil {
				metaValues.Values = append(metaValues.Values, metaValue)
			}
		}
	}

	for key, value := range request.Form {
		newMetaValue(key, value)
	}

	if request.MultipartForm != nil {
		for key, value := range request.MultipartForm.File {
			newMetaValue(key, value)
		}
	}
	return metaValues, nil
}

func Decode(contextor qor.Contextor, result interface{}, res Resourcer) (errs []error) {
	context := contextor.GetContext()
	var err error
	var metaValues *MetaValues
	metaors := res.GetMetas()

	responder.With("html", func() {
		metaValues, err = ConvertFormToMetaValues(context.Request, metaors, "QorResource.")
	}).With("json", func() {
		metaValues, err = ConvertJSONToMetaValues(context.Request.Body, metaors)
		context.Request.Body.Close()
	}).Respond(nil, context.Request)

	errs = DecodeToResource(res, result, metaValues, context).Start()
	return errs
}

func getAddrValue(value reflect.Value) interface{} {
	if value.Kind() == reflect.Ptr {
		return value.Interface()
	} else if value.CanAddr() {
		return value.Addr().Interface()
	} else {
		return value.Interface()
	}
}

func ConvertObjectToMap(contextor qor.Contextor, object interface{}, metaors []Metaor) interface{} {
	context := contextor.GetContext()
	reflectValue := reflect.Indirect(reflect.ValueOf(object))

	switch reflectValue.Kind() {
	case reflect.Slice:
		values := []interface{}{}
		for i := 0; i < reflectValue.Len(); i++ {
			values = append(values, ConvertObjectToMap(context, getAddrValue(reflectValue.Index(i)), metaors))
		}
		return values
	case reflect.Struct:
		values := map[string]interface{}{}
		for _, metaor := range metaors {
			if metaor.HasPermission(roles.Read, context) {
				value := metaor.GetValuer()(object, context)
				if len(metaor.GetMetas()) > 0 {
					value = ConvertObjectToMap(context, value, metaor.GetMetas())
				}
				values[metaor.GetName()] = value
			}
		}
		return values
	default:
		panic(fmt.Sprintf("Can't convert %v (%v) to map", reflectValue, reflectValue.Kind()))
	}
}
