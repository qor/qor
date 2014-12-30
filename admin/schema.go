package admin

import (
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"

	"reflect"
	"regexp"
	"strings"
)

type Resourcer interface {
	AllMetas() []*resource.Meta
}

func ConvertMapToMetaValues(values map[string]interface{}, res Resourcer) (metaValues *resource.MetaValues) {
	metas := make(map[string]resource.Metaor)
	if res != nil {
		for _, attr := range res.AllMetas() {
			metas[attr.Name] = attr
		}
	}

	metaValues = new(resource.MetaValues)
	for key, value := range values {
		meta := metas[key]
		if str, ok := value.(string); ok {
			metaValue := &resource.MetaValue{Name: key, Value: str, Meta: meta}
			metaValues.Values = append(metaValues.Values, metaValue)
		} else {
			var res Resourcer
			if meta != nil && meta.GetMeta() != nil && meta.GetMeta().Resource != nil {
				res, _ = meta.GetMeta().Resource.(Resourcer)
			}

			if vs, ok := value.(map[string]interface{}); ok {
				children := ConvertMapToMetaValues(vs, res)
				metaValue := &resource.MetaValue{Name: key, Meta: meta, MetaValues: children}
				metaValues.Values = append(metaValues.Values, metaValue)
			} else if vs, ok := value.([]interface{}); ok {
				for _, v := range vs {
					if mv, ok := v.(map[string]interface{}); ok {
						children := ConvertMapToMetaValues(mv, res)
						metaValue := &resource.MetaValue{Name: key, Meta: meta, MetaValues: children}
						metaValues.Values = append(metaValues.Values, metaValue)
					} else if meta != nil {
						metaValue := &resource.MetaValue{Name: key, Value: vs, Meta: meta}
						metaValues.Values = append(metaValues.Values, metaValue)
						break
					}
				}
			} else {
				switch reflect.ValueOf(value).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Bool:
					metaValue := &resource.MetaValue{Name: key, Value: value, Meta: meta}
					metaValues.Values = append(metaValues.Values, metaValue)
				default:
					panic("doesn't support this type:" + reflect.ValueOf(value).Kind().String())
				}
			}
		}
	}
	return
}

func ConvertFormToMetaValues(context *Context, prefix string, res *Resource) (metaValues *resource.MetaValues) {
	request := context.Request
	convertedMap := make(map[string]bool)
	metas := make(map[string]resource.Metaor)
	if res != nil {
		for _, attr := range res.AllMetas() {
			metas[attr.Name] = attr
		}
	}

	metaValues = new(resource.MetaValues)
	for key := range request.Form {
		if strings.HasPrefix(key, prefix) {
			key = strings.TrimPrefix(key, prefix)
			isCurrent := regexp.MustCompile("^[^.]+$")
			isNext := regexp.MustCompile(`^(([^.\[\]]+)(\[\d+\])?)(?:\.([^.]+)+)$`)

			if matches := isCurrent.FindStringSubmatch(key); len(matches) > 0 {
				meta := metas[matches[0]]
				metaValue := &resource.MetaValue{Name: matches[0], Value: request.Form[prefix+key], Meta: meta}
				metaValues.Values = append(metaValues.Values, metaValue)
			} else if matches := isNext.FindStringSubmatch(key); len(matches) > 0 {
				if _, ok := convertedMap[matches[1]]; !ok {
					convertedMap[matches[1]] = true
					meta := metas[matches[2]]
					var res *Resource
					if meta != nil && meta.GetMeta() != nil {
						res = meta.GetMeta().Resource.(*Resource)
					}
					children := ConvertFormToMetaValues(context, prefix+matches[1]+".", res)
					metaValue := &resource.MetaValue{Name: matches[2], Meta: meta, MetaValues: children}
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

func ConvertObjectToMap(context *Context, object interface{}, res *Resource) interface{} {
	reflectValue := reflect.Indirect(reflect.ValueOf(object))
	switch reflectValue.Kind() {
	case reflect.Slice:
		len := reflectValue.Len()
		values := []interface{}{}
		for i := 0; i < len; i++ {
			values = append(values, ConvertObjectToMap(context, reflectValue.Index(i).Interface(), res))
		}
		return values
	case reflect.Struct:
		values := map[string]interface{}{}
		metas := res.ShowMetas()
		for _, meta := range metas {
			if meta.HasPermission(roles.Read, context.Context) {
				value := meta.Value(object, context.Context)
				if res, ok := meta.Resource.(*Resource); ok {
					value = ConvertObjectToMap(context, value, res)
				}
				values[meta.Name] = value
			}
		}
		return values
	default:
		panic("can't convert object to map")
	}
}
