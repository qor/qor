package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/responder"
	"github.com/qor/qor/roles"
)

type Contextor interface {
	GetContext() *qor.Context
}

func (res *Resource) ConvertMapToMetaValues(values map[string]interface{}) (metaValues *MetaValues) {
	metas := make(map[string]Metaor)
	if res != nil {
		for _, attr := range res.GetMetas() {
			metas[attr.Name] = attr
		}
	}

	metaValues = new(MetaValues)
	for key, value := range values {
		meta := metas[key]
		if str, ok := value.(string); ok {
			metaValue := &MetaValue{Name: key, Value: str, Meta: meta}
			metaValues.Values = append(metaValues.Values, metaValue)
		} else {
			var res *Resource
			if meta != nil && meta.GetMeta() != nil && meta.GetMeta().Resource != nil {
				res = meta.GetMeta().Resource.GetResource()
			}

			if vs, ok := value.(map[string]interface{}); ok {
				children := res.ConvertMapToMetaValues(vs)
				metaValue := &MetaValue{Name: key, Meta: meta, MetaValues: children}
				metaValues.Values = append(metaValues.Values, metaValue)
			} else if vs, ok := value.([]interface{}); ok {
				for _, v := range vs {
					if mv, ok := v.(map[string]interface{}); ok {
						children := res.ConvertMapToMetaValues(mv)
						metaValue := &MetaValue{Name: key, Meta: meta, MetaValues: children}
						metaValues.Values = append(metaValues.Values, metaValue)
					} else if meta != nil {
						metaValue := &MetaValue{Name: key, Value: vs, Meta: meta}
						metaValues.Values = append(metaValues.Values, metaValue)
						break
					}
				}
			} else {
				switch reflect.ValueOf(value).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
					metaValue := &MetaValue{Name: key, Value: value, Meta: meta}
					metaValues.Values = append(metaValues.Values, metaValue)
				default:
					panic("doesn't support this type:" + reflect.ValueOf(value).Kind().String())
				}
			}
		}
	}
	return
}

func (res *Resource) ConvertFormToMetaValues(contextor Contextor, prefix string) (metaValues *MetaValues) {
	context := contextor.GetContext()
	request := context.Request
	convertedMap := make(map[string]bool)
	metas := make(map[string]Metaor)
	if res != nil {
		for _, attr := range res.GetMetas() {
			metas[attr.Name] = attr
		}
	}

	metaValues = new(MetaValues)
	for key := range request.Form {
		if strings.HasPrefix(key, prefix) {
			key = strings.TrimPrefix(key, prefix)
			isCurrent := regexp.MustCompile("^[^.]+$")
			isNext := regexp.MustCompile(`^(([^.\[\]]+)(\[\d+\])?)(?:\.([^.]+)+)$`)

			if matches := isCurrent.FindStringSubmatch(key); len(matches) > 0 {
				meta := metas[matches[0]]
				metaValue := &MetaValue{Name: matches[0], Value: request.Form[prefix+key], Meta: meta}
				metaValues.Values = append(metaValues.Values, metaValue)
			} else if matches := isNext.FindStringSubmatch(key); len(matches) > 0 {
				if _, ok := convertedMap[matches[1]]; !ok {
					convertedMap[matches[1]] = true
					meta := metas[matches[2]]
					var res *Resource
					if meta != nil && meta.GetMeta() != nil {
						res = meta.GetMeta().Resource.GetResource()
					}
					children := res.ConvertFormToMetaValues(context, prefix+matches[1]+".")
					metaValue := &MetaValue{Name: matches[2], Meta: meta, MetaValues: children}
					metaValues.Values = append(metaValues.Values, metaValue)
				}
			}
		}
	}

	if request.MultipartForm != nil {
		// for key, header := range request.MultipartForm.File {
		// xxxxx
		// }
	}
	return
}

func (res *Resource) ConvertObjectToMap(contextor Contextor, object interface{}) interface{} {
	context := contextor.GetContext()
	reflectValue := reflect.Indirect(reflect.ValueOf(object))
	switch reflectValue.Kind() {
	case reflect.Slice:
		len := reflectValue.Len()
		values := []interface{}{}
		for i := 0; i < len; i++ {
			values = append(values, res.ConvertObjectToMap(context, reflectValue.Index(i).Interface()))
		}
		return values
	case reflect.Struct:
		values := map[string]interface{}{}
		metas := res.GetMetas()
		for _, meta := range metas {
			if meta.HasPermission(roles.Read, context) {
				value := meta.Value(object, context)
				if res, ok := meta.Resource.(*Resource); ok {
					value = res.ConvertObjectToMap(context, value)
				}
				values[meta.Name] = value
			}
		}
		return values
	default:
		panic(fmt.Sprintf("Can't convert %v (%v) to map", object, reflectValue.Kind()))
	}
}

func (res *Resource) Decode(contextor Contextor, result interface{}) (errs []error) {
	context := contextor.GetContext()
	responder.With("html", func() {
		errs = DecodeToResource(res, result, res.ConvertFormToMetaValues(context, "QorResource."), context).Start()
	}).With("json", func() {
		decoder := json.NewDecoder(context.Request.Body)
		values := map[string]interface{}{}
		if err := decoder.Decode(&values); err == nil {
			errs = DecodeToResource(res, result, res.ConvertMapToMetaValues(values), context).Start()
		} else {
			errs = append(errs, err)
		}
	}).Respond(nil, context.Request)
	return errs
}
