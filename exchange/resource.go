package exchange

import (
	"fmt"
	"strings"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Resource struct {
	*resource.Resource
	Metas                map[string]*Meta // TODO: replace it with a slice
	AutoCreate           bool
	MultiDelimiter       string
	HasSequentialColumns bool
	HeadersInOrder       []string
}

func NewResource(val interface{}) *Resource {
	res := &Resource{
		Resource:   &resource.Resource{Value: val},
		AutoCreate: true,
		Metas:      map[string]*Meta{},
	}
	res.AddValidator(func(_ interface{}, mvs *resource.MetaValues, ctx *qor.Context) error {
		for _, meta := range res.Metas {
			// if meta, ok := mr.(*Meta); ok {
			var hasMeta bool

			for _, mv := range mvs.Values {
				if mv.Name == meta.Name {
					hasMeta = true
					break
				}
			}
			if !hasMeta && !meta.Optional && meta.Resource == nil {
				return fmt.Errorf("exchange: should contains Meta %s in MetaValues", meta.Name)
			}
			// }
		}

		return nil
	})

	return res
}

func (res *Resource) CallFinder(result interface{}, metaValues *resource.MetaValues, ctx *qor.Context) (err error) {
	if res.Finder != nil {
		err = res.Finder(result, metaValues, ctx)
		if err == resource.ErrProcessorRecordNotFound && res.AutoCreate {
			err = nil
		}
	} else if !res.AutoCreate {
		err = resource.ErrProcessorRecordNotFound
	}

	return
}

func (res *Resource) Meta(meta *Meta) *Meta {
	// m := &Meta{Meta: meta}
	// res.Resource.Meta(m)
	meta.base = res
	meta.updateMeta()
	// res.Metas = append(res.Metas, meta)
	res.Metas[meta.Name] = meta
	res.HeadersInOrder = append(res.HeadersInOrder, meta.Name)
	return meta
}

func (res *Resource) getMetaValues(vmap map[string]string, index int) (mvs *resource.MetaValues, validatedIndex bool) {
	mvs = new(resource.MetaValues)
	for _, m := range res.Metas {
		// m, ok := mr.(*Meta)
		// if !ok {
		// 	continue
		// }

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
			for _, subVmap := range metaResource.getSubVmaps(vmap) {
				subMvs, _ := metaResource.getMetaValues(subVmap, 0)
				mvs.Values = append(mvs.Values, &resource.MetaValue{
					Name:       m.Name,
					Meta:       m,
					MetaValues: subMvs,
				})
			}
		} else {
			mv.MetaValues, _ = metaResource.getMetaValues(vmap, index)
			mvs.Values = append(mvs.Values, &mv)
		}
	}

	return
}

func (res *Resource) getSubVmaps(vmap map[string]string) (subVmaps []map[string]string) {
	for _, meta := range res.Metas {
		for k, v := range vmap {
			// meta := metaor.GetMeta()
			if meta.Label == k {
				for i, subv := range strings.Split(v, ",") {
					if len(subVmaps) == i {
						subVmaps = append(subVmaps, make(map[string]string))
					}
					subVmaps[i][k] = strings.TrimSpace(subv)
				}
			} else if meta.Resource != nil {
				subResource, ok := meta.Resource.(*Resource)
				if !ok {
					continue
				}
				subMetaSubVmaps := subResource.getSubVmaps(vmap)
				for i, subMetaVmap := range subMetaSubVmaps {
					if len(subVmaps) == i {
						subVmaps = append(subVmaps, make(map[string]string))
					}
					vmap := subVmaps[i]
					for k, v := range subMetaVmap {
						vmap[k] = v
					}
				}
			}
		}
	}

	return
}
