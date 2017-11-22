package resource

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/utils"
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
		var childMeta []Metaor
		if metaor != nil {
			childMeta = metaor.GetMetas()
		}

		switch result := value.(type) {
		case map[string]interface{}:
			if children, err := convertMapToMetaValues(result, childMeta); err == nil {
				metaValue = &MetaValue{Name: key, Meta: metaor, MetaValues: children}
			}
		case []interface{}:
			for idx, r := range result {
				if mr, ok := r.(map[string]interface{}); ok {
					if children, err := convertMapToMetaValues(mr, childMeta); err == nil {
						metaValue := &MetaValue{Name: key, Meta: metaor, MetaValues: children, Index: idx}
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

// ConvertJSONToMetaValues convert json to meta values
func ConvertJSONToMetaValues(reader io.Reader, metaors []Metaor) (*MetaValues, error) {
	var (
		err     error
		values  = map[string]interface{}{}
		decoder = json.NewDecoder(reader)
	)

	if err = decoder.Decode(&values); err == nil {
		return convertMapToMetaValues(values, metaors)
	}
	return nil, err
}

var (
	isCurrentLevel = regexp.MustCompile("^[^.]+$")
	isNextLevel    = regexp.MustCompile(`^(([^.\[\]]+)(\[\d+\])?)(?:(\.[^.]+)+)$`)
)

// ConvertFormToMetaValues convert form to meta values
func ConvertFormToMetaValues(request *http.Request, metaors []Metaor, prefix string) (*MetaValues, error) {
	metaValues := &MetaValues{}
	metaorsMap := map[string]Metaor{}
	convertedNextLevel := map[string]bool{}
	nestedStructIndex := map[string]int{}
	for _, metaor := range metaors {
		metaorsMap[metaor.GetName()] = metaor
	}

	newMetaValue := func(key string, value interface{}) {
		if strings.HasPrefix(key, prefix) {
			var metaValue *MetaValue
			key = strings.TrimPrefix(key, prefix)

			if matches := isCurrentLevel.FindStringSubmatch(key); len(matches) > 0 {
				name := matches[0]
				metaValue = &MetaValue{Name: name, Meta: metaorsMap[name], Value: value}
			} else if matches := isNextLevel.FindStringSubmatch(key); len(matches) > 0 {
				name := matches[1]
				if _, ok := convertedNextLevel[name]; !ok {
					var metaors []Metaor
					convertedNextLevel[name] = true
					metaor := metaorsMap[matches[2]]
					if metaor != nil {
						metaors = metaor.GetMetas()
					}

					if children, err := ConvertFormToMetaValues(request, metaors, prefix+name+"."); err == nil {
						nestedName := prefix + matches[2]
						if _, ok := nestedStructIndex[nestedName]; ok {
							nestedStructIndex[nestedName]++
						} else {
							nestedStructIndex[nestedName] = 0
						}

						// is collection
						if matches[3] != "" {
							metaValue = &MetaValue{Name: matches[2], Meta: metaor, MetaValues: children, Index: nestedStructIndex[nestedName]}
						} else {
							// is nested and it is existing
							if metaValue = metaValues.Get(matches[2]); metaValue == nil {
								metaValue = &MetaValue{Name: matches[2], Meta: metaor, MetaValues: children, Index: nestedStructIndex[nestedName]}
							} else {
								metaValue.MetaValues = children
								metaValue.Index = nestedStructIndex[nestedName]
								metaValue = nil
							}
						}
					}
				}
			}

			if metaValue != nil {
				metaValues.Values = append(metaValues.Values, metaValue)
			}
		}
	}

	var sortedFormKeys []string
	for key := range request.Form {
		sortedFormKeys = append(sortedFormKeys, key)
	}

	utils.SortFormKeys(sortedFormKeys)

	for _, key := range sortedFormKeys {
		newMetaValue(key, request.Form[key])
	}

	if request.MultipartForm != nil {
		sortedFormKeys = []string{}
		for key := range request.MultipartForm.File {
			sortedFormKeys = append(sortedFormKeys, key)
		}
		utils.SortFormKeys(sortedFormKeys)

		for _, key := range sortedFormKeys {
			newMetaValue(key, request.MultipartForm.File[key])
		}
	}
	return metaValues, nil
}

// Decode decode context to result according to resource definition
func Decode(context *qor.Context, result interface{}, res Resourcer) error {
	var errors qor.Errors
	var err error
	var metaValues *MetaValues
	metaors := res.GetMetas([]string{})

	if strings.Contains(context.Request.Header.Get("Content-Type"), "json") {
		metaValues, err = ConvertJSONToMetaValues(context.Request.Body, metaors)
		context.Request.Body.Close()
	} else {
		metaValues, err = ConvertFormToMetaValues(context.Request, metaors, "QorResource.")
	}

	errors.AddError(err)
	errors.AddError(DecodeToResource(res, result, metaValues, context).Start())
	return errors
}
