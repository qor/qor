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

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
	"github.com/tealeg/xlsx"
)

// TODO: support csv files

type Exchange struct {
	Resource *Resource

	// TODO
	JobThrottle      int
	StopOnError      bool
	NormalizeHeaders func(sheet *xlsx.Sheet) []string
}

func New(res *Resource) *Exchange {
	return &Exchange{
		Resource:    res,
		JobThrottle: 1,
		NormalizeHeaders: func(sheet *xlsx.Sheet) (headers []string) {
			for _, c := range sheet.Rows[0].Cells {
				headers = append(headers, c.Value)
			}
			return
		},
	}
}

type Resource struct {
	resource.Resource

	// TODO
	AutoCreate           bool
	MultiDelimiter       string
	HasSequentialColumns bool
}

func NewResource(val interface{}) *Resource {
	return &Resource{Resource: resource.Resource{Value: val}}
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

func (ex *Exchange) Import(r io.Reader, ctx *qor.Context) (fileInfo FileInfo, importStatusChan chan ImportStatus, err error) {
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

	fileInfo.TotalLines, xf = preprocessXLSXFile(xf)
	fileInfo.Done = make(chan bool)
	fileInfo.Error = make(chan error)
	importStatusChan = make(chan ImportStatus, 10)

	go ex.process(xf, ctx, fileInfo, importStatusChan)

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

func (ex *Exchange) process(xf *xlsx.File, ctx *qor.Context, fileInfo FileInfo, importStatusChan chan ImportStatus) {
	var wait sync.WaitGroup
	wait.Add(fileInfo.TotalLines - 1)
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
	res := ex.Resource
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
				importStatus := ImportStatus{Sheet: sheet.Name}
				defer func() {
					setError(len(importStatus.Errors) > 0)
					importStatusChan <- importStatus
					<-throttle
					wait.Done()
				}()

				vmap := map[string]string{}
				for j, cell := range row.Cells {
					vmap[headers[j].Value] = cell.Value
				}

				importStatus.MetaValues = res.getMetaValues(vmap, 0)
				processor := resource.DecodeToResource(res, res.NewStruct(), importStatus.MetaValues, ctx)

				// TODO: handle skip left
				if err := processor.Initialize(); err != nil && err != resource.ErrProcessorRecordNotFound {
					importStatus.Errors = []error{err}
					return
				}

				if errs := processor.Validate(); len(errs) > 0 {
					importStatus.Errors = errs
					return
				}

				if errs := processor.Commit(); len(errs) > 0 {
					importStatus.Errors = errs
					return
				}

				// TODO: replace it with processor.Save
				if err := db.Save(processor.Result).Error; err != nil {
					importStatus.Errors = []error{err}
					return
				}
			}(i, row, importStatusChan)
		}
	}

	wait.Wait()

	if hasError {
		goto rollback
	}

	if err := db.Commit().Error; err != nil {
		fileInfo.Error <- err
		return
	}
	fileInfo.Done <- true
	return

rollback:
	if err := db.Rollback().Error; err != nil {
		fileInfo.Error <- err
	}
	fileInfo.Error <- errors.New("meet error in job processing")
	return
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
