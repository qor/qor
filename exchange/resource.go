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

func (res *Resource) RegisterMeta(meta *resource.Meta) *Meta {
	m := &Meta{Meta: meta}
	res.Resource.RegisterMeta(m)
	return m
}

func (res *Resource) getMetaValues(vmap map[string]string, index int) (mvs *resource.MetaValues) {
	mvs = new(resource.MetaValues)
	for _, mr := range res.Metas {
		m, ok := mr.(*Meta)
		if !ok {
			continue
		}

		label := m.Label
		if index > 0 {
			label = fmtLabel(label, index)
		}

		mv := resource.MetaValue{Name: m.Name, Meta: m}
		if m.Resource == nil {
			mv.Value = vmap[label]
			delete(vmap, label)
			mvs.Values = append(mvs.Values, &mv)

			continue
		}

		metaResource, ok := m.Resource.(*Resource)
		if !ok {
			continue
		}

		if !metaResource.HasSequentialColumns {
			mv.MetaValues = metaResource.getMetaValues(vmap, 0)
			mvs.Values = append(mvs.Values, &mv)

			continue
		}

		i := 1
		markMeta := metaResource.getNonOptionalMeta()
		for {
			if _, ok := vmap[fmtLabel(markMeta.Label, i)]; !ok {
				break
			}

			nmv := mv
			nmv.MetaValues = metaResource.getMetaValues(vmap, i)
			mvs.Values = append(mvs.Values, &nmv)
			i++
		}
	}

	return
}

// TODO: support both "header 01" and "header 1"
func fmtLabel(l string, i int) string {
	return fmt.Sprintf("%s %#02d", l, i)
}

// TODO: what if all metas are optional?
func (res *Resource) getNonOptionalMeta() *Meta {
	for _, mr := range res.Metas {
		m, ok := mr.(*Meta)
		if !ok {
			continue
		}

		if !m.Optional {
			return m
		}
	}

	return nil
}
