package exchange

import (
	"fmt"
	"reflect"

	"github.com/qor/qor/resource"
)

type Resource struct {
	resource.Resource

	// TODO
	AlwaysCreate         bool
	AutoCreate           bool
	MultiDelimiter       string
	HasSequentialColumns bool
}

func NewResource(val interface{}) *Resource {
	return &Resource{Resource: resource.Resource{Value: val}}
}

type Meta struct {
	*resource.Meta

	// TODO
	Optional     bool // make use of validator?
	AliasHeaders []string
}

func (m *Meta) Set(field string, val interface{}) *Meta {
	reflect.ValueOf(m).Elem().FieldByName(field).Set(reflect.ValueOf(val))
	return m
}

func (m *Meta) getCurrentLabel(vmap map[string]string, index int) string {
	var labels []string
	if index > 0 {
		// support both "label 01" and "label 1"
		labels = append(labels, fmt.Sprintf("%s %#02d", m.Label, index), fmt.Sprintf("%s %d", m.Label, index))
	} else {
		labels = append(labels, m.Label)
	}

	labels = append(labels, m.AliasHeaders...)
	for _, label := range labels {
		if _, ok := vmap[label]; ok {
			return label
		}
	}

	return ""
}

func (res *Resource) RegisterMeta(meta *resource.Meta) *Meta {
	m := &Meta{Meta: meta}
	res.Resource.RegisterMeta(m)
	return m
}

func (res *Resource) getMetaValues(vmap map[string]string, index int) (mvs *resource.MetaValues, validatedIndex bool) {
	mvs = new(resource.MetaValues)
	for _, mr := range res.Metas {
		m, ok := mr.(*Meta)
		if !ok {
			continue
		}
		mv := resource.MetaValue{Name: m.Name, Meta: m}
		if m.Resource == nil {
			if label := m.getCurrentLabel(vmap, index); label != "" {
				mv.Value = vmap[label]
				delete(vmap, label)
				mvs.Values = append(mvs.Values, &mv)
				validatedIndex = true
			}

			continue
		}
		metaResource, ok := m.Resource.(*Resource)
		if !ok {
			continue
		}
		if metaResource.HasSequentialColumns {
			for i := 1; ; i++ {
				subMvs, validate := metaResource.getMetaValues(vmap, i)
				if !validate {
					break
				}

				validatedIndex = true
				mvs.Values = append(mvs.Values, &resource.MetaValue{
					Name:       m.Name,
					Meta:       m,
					MetaValues: subMvs,
				})
			}
		} else if metaResource.MultiDelimiter != "" {
		} else {
			mv.MetaValues, _ = metaResource.getMetaValues(vmap, 0)
			mvs.Values = append(mvs.Values, &mv)
		}
	}

	return
}
