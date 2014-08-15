package exchange

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/tealeg/xlsx"
)

type Exchange struct {
	Resources []*Resource
	DB        *gorm.DB
}

func New(db *gorm.DB) *Exchange {
	return &Exchange{DB: db}
}

func (e *Exchange) NewResource(val interface{}) *Resource {
	res := &Resource{Resource: resource.Resource{Value: val}}
	e.Resources = append(e.Resources, res)
	return res
}

type Resource struct {
	resource.Resource

	// TODO
	AutoCreate  bool
	StopOnError bool
	JobThrottle int

	MultiDelimiter       string
	HasSequentialColumns bool
}

func (res *Resource) RegisterMeta(meta *resource.Meta) *Meta {
	m := &Meta{Meta: meta}
	res.Resource.RegisterMeta(m)
	return m
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

// TODO
func (m *Meta) NormalizeHeaders() {

}

func ImportFileName() {}
func ImportFile()     {}

type ImportStatus struct {
	LineNum    int
	Sheet      string
	MetaValues *resource.MetaValues
	Errors     []error
}

type FileInfo struct {
	TotalLines int
	Done       chan bool
	Error      chan error
}

func (res *Resource) Import(r io.Reader, ctx *qor.Context) (fi FileInfo, iic chan ImportStatus, err error) {
	f, err := ioutil.TempFile("", "qor.exchange.")
	if err != nil {
		return
	}
	defer func() { f.Close() }()
	_, err = io.Copy(f, r)
	if err != nil {
		return
	}
	defer func() { os.Remove(f.Name()) }()

	zr, err := zip.OpenReader(f.Name())
	if err != nil {
		return
	}
	xf, err := xlsx.ReadZip(zr)
	if err != nil {
		return
	}

	fi.TotalLines, xf = preprocessXLSXFile(xf)
	fi.Done = make(chan bool)
	fi.Error = make(chan error)
	iic = make(chan ImportStatus, 10)

	go res.process(xf, ctx, fi, iic)

	return
}

func preprocessXLSXFile(xf *xlsx.File) (totalLines int, nxf *xlsx.File) {
	nxf = new(xlsx.File)
	for _, sheet := range xf.Sheets {
		if len(sheet.Rows) == 0 {
			continue
		}

		nsheet := *sheet
		nsheet.Rows = []*xlsx.Row{}
		for _, row := range sheet.Rows {
			if len(row.Cells) == 0 {
				continue
			}

			empty := true
			for _, cell := range row.Cells {
				if cell.Value == "" {
					continue
				}

				empty = false
				break
			}

			if empty {
				continue
			}

			nsheet.Rows = append(nsheet.Rows, row)
		}

		nsheet.MaxRow = len(nsheet.Rows)
		totalLines += nsheet.MaxRow
		nxf.Sheets = append(nxf.Sheets, &nsheet)
	}

	return
}

func (res *Resource) process(xf *xlsx.File, ctx *qor.Context, fi FileInfo, iic chan ImportStatus) {
	var wait sync.WaitGroup
	wait.Add(fi.TotalLines - 1)
	throttle := make(chan bool, 20)
	defer func() { close(throttle) }()
	var hasError bool
	lock := new(sync.Mutex)
	setError := func(h bool) {
		lock.Lock()
		if !hasError {
			hasError = h
		}
		lock.Unlock()
	}

	db := ctx.DB.Begin()
	for _, sheet := range xf.Sheets {
		if len(sheet.Rows) <= 1 {
			continue
		}

		headers := sheet.Rows[0].Cells
		for i, row := range sheet.Rows[1:] {
			throttle <- true
			if hasError {
				goto rollback
			}

			go func(line int, row *xlsx.Row, iic chan ImportStatus) {
				ii := ImportStatus{Sheet: sheet.Name}
				defer func() {
					setError(len(ii.Errors) > 0)
					iic <- ii
					<-throttle
					wait.Done()
				}()

				vmap := map[string]string{}
				for j, cell := range row.Cells {
					vmap[headers[j].Value] = cell.Value
				}

				ii.MetaValues = res.GetMetaValues(vmap, 0)
				p := resource.DecodeToResource(res, res.NewStruct(), ii.MetaValues, ctx)

				if err := p.Initialize(); err != nil && err != resource.ErrProcessorRecordNotFound {
					ii.Errors = []error{err}
					return
				}

				if errs := p.Validate(); len(errs) > 0 {
					ii.Errors = errs
					return
				}

				if errs := p.Commit(); len(errs) > 0 {
					ii.Errors = errs
					return
				}

				if err := db.Save(p.Result).Error; err != nil {
					ii.Errors = []error{err}
					return
				}
			}(i, row, iic)
		}
	}

	wait.Wait()

	if hasError {
		goto rollback
	}

	if err := db.Commit().Error; err != nil {
		fi.Error <- err
		return
	}
	fi.Done <- true
	return

rollback:
	if err := db.Rollback().Error; err != nil {
		// log.Println("exchange: rollback:", err.Error())
		fi.Error <- err
	}
	fi.Error <- errors.New("meet error in job processing")
	return
}

// // TODO: should handle this in package resource?
// func formatErrors(line int, errs []error) error {
// 	var msg string
// 	for _, e := range errs {
// 		msg += e.Error() + ";"
// 	}

// 	return fmt.Errorf("line %d: %s", line, msg)
// }

func (res *Resource) GetMetaValues(vmap map[string]string, index int) (mvs *resource.MetaValues) {
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

		mv := resource.MetaValue{Name: label, Meta: m}
		if m.Resource == nil {
			mv.Value = vmap[label]
			delete(vmap, label)
			mvs.Values = append(mvs.Values, &mv)

			continue
		}

		ms, ok := m.Resource.(*Resource)
		if !ok {
			continue
		}

		if !ms.HasSequentialColumns {
			mv.MetaValues = ms.GetMetaValues(vmap, 0)
			mvs.Values = append(mvs.Values, &mv)

			continue
		}

		i := 1
		markMeta := ms.getNonOptionalMeta()
		for {
			if _, ok := vmap[fmtLabel(markMeta.Label, i)]; !ok {
				break
			}

			nmv := mv
			nmv.MetaValues = ms.GetMetaValues(vmap, i)
			mvs.Values = append(mvs.Values, &nmv)
			i++
		}
	}

	return
}

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
