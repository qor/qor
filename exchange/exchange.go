package exchange

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/tealeg/xlsx"
)

type Exchange struct {
	Resources []*Resource
	DB        *gorm.DB
}

func (e *Exchange) NewResource(val interface{}) *Resource {
	res := &Resource{Resource: resource.Resource{Value: val}}
	e.Resources = append(e.Resources, res)
	return res
}

type Resource struct {
	resource.Resource
}

type Meta struct {
	resource.Meta
	MultiDelimiter       string
	HasSequentialColumns bool
}

func ImportFileName() {}
func ImportFile()     {}

func (res *Resource) Import(r io.Reader, ctx *qor.Context) (err error) {
	f, err := ioutil.TempFile("", "qor.exchange.")
	if err != nil {
		return errors.New("exchange: " + err.Error())
	}
	defer func() { f.Close() }()
	_, err = io.Copy(f, r)
	if err != nil {
		return errors.New("exchange: " + err.Error())
	}
	defer func() { os.Remove(f.Name()) }()

	zr, err := zip.OpenReader(f.Name())
	if err != nil {
		return errors.New("exchange: " + err.Error())
	}
	xf, err := xlsx.ReadZip(zr)
	if err != nil {
		return errors.New("exchange: " + err.Error())
	}

	// var mds resource.MetaDatas
	ctx.DB.Begin()
	for _, sheet := range xf.Sheets {
		if len(sheet.Rows) <= 1 {
			continue
		}

		headers := sheet.Rows[0].Cells
		for i, row := range sheet.Rows[1:] {
			vmap := map[string]string{}
			empty := true
			for j, cell := range row.Cells {
				vmap[headers[j].Value] = cell.Value
				if empty {
					empty = cell.Value == ""
				}
			}
			if empty {
				continue
			}

			mds := res.GetMetaDatas(vmap)
			p := resource.DecodeToResource(res, res.NewStruct(), mds, ctx)
			err = p.Initialize()
			if err != nil && err != resource.ErrProcessorRecordNotFound {
				err = formatErrors(i+1, []error{err})
				break
			}
			err = nil
			errs := p.Validate()
			if len(errs) > 0 {
				err = formatErrors(i+1, errs)
				break
			}
			errs = p.Commit()
			if len(errs) > 0 {
				err = formatErrors(i+1, errs)
				break
			}
			ctx.DB.Save(p.Result)
		}
	}
	if err != nil {
		ctx.DB.Rollback()
	} else {
		ctx.DB.Commit()
	}

	return
}

// TODO: should handle this in package resource?
func formatErrors(line int, errs []error) error {
	var msg string
	for _, e := range errs {
		msg += e.Error() + ";"
	}

	return fmt.Errorf("line %d: %s", line, msg)
}

func (res *Resource) GetMetaDatas(vmap map[string]string) (mds resource.MetaDatas) {
	for _, mr := range res.Metas {
		m, ok := mr.(*Meta)
		if !ok {
			continue
		}

		md := resource.MetaData{Name: m.Label, Meta: m}
		if m.Resource == nil {
			md.Value = vmap[m.Label]
			delete(vmap, m.Label)
			mds = append(mds, &md)

			continue
		}

		ms, ok := m.Resource.(*Resource)
		if !ok {
			continue
		}

		md.MetaDatas = ms.GetMetaDatas(vmap)
		mds = append(mds, &md)
	}

	return
}
