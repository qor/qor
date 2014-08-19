package exchange

import (
	"errors"
	"sync"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

// TODO: support csv files
// TODO: support ImportStatus chan data sorter

type Exchange struct {
	Resource *Resource

	// TODO
	StopOnError bool

	JobThrottle      int
	StatusThrottle   int
	NormalizeHeaders func(f File) []string
	DataStartAt      int
}

func New(res *Resource) *Exchange {
	return &Exchange{
		Resource:       res,
		JobThrottle:    1,
		DataStartAt:    1,
		StatusThrottle: 10,
		NormalizeHeaders: func(f File) (headers []string) {
			if f.TotalLines() <= 0 {
				return
			}

			return f.Line(0)
		},
	}
}

func ImportFileName() {}
func ImportFile()     {}

type ImportStatus struct {
	LineNum    int
	MetaValues *resource.MetaValues
	Errors     []error
}

type FileInfo struct {
	TotalLines int
	Done       chan bool
	Error      chan error
}

type File interface {
	TotalLines() (num int)
	Line(l int) (fields []string)
}

func (ex *Exchange) Import(f File, ctx *qor.Context) (fileInfo FileInfo, importStatusChan chan ImportStatus, err error) {
	fileInfo.TotalLines = f.TotalLines()
	fileInfo.Done = make(chan bool)
	fileInfo.Error = make(chan error)
	importStatusChan = make(chan ImportStatus, ex.StatusThrottle)

	go ex.process(f, ex.NormalizeHeaders(f), ctx, fileInfo, importStatusChan)

	return
}

func (ex *Exchange) process(f File, headers []string, ctx *qor.Context, fileInfo FileInfo, importStatusChan chan ImportStatus) {
	var wait sync.WaitGroup
	wait.Add(fileInfo.TotalLines - ex.DataStartAt)
	throttle := make(chan bool, ex.JobThrottle)
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
	for num := ex.DataStartAt; num < fileInfo.TotalLines; num++ {
		throttle <- true
		if hasError {
			goto rollback
		}

		go func(num int, importStatusChan chan ImportStatus) {
			var importStatus ImportStatus
			defer func() {
				setError(len(importStatus.Errors) > 0)

				<-throttle
				importStatusChan <- importStatus
				wait.Done()
			}()

			vmap := map[string]string{}
			line := f.Line(num)
			lineLen := len(line)
			for j, header := range headers {
				if j >= lineLen {
					break
				}

				vmap[header] = line[j]
			}

			importStatus.MetaValues, _ = res.getMetaValues(vmap, 0)
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
		}(num, importStatusChan)
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
