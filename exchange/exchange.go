package exchange

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
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

				// can't replace this with resource.CallSafer for the sake of transaction
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
