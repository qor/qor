package exchange

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

// TODO: support csv files
// TODO: support ImportStatus chan data sorter

type Exchange struct {
	Resource         *Resource
	StopOnError      bool
	JobThrottle      int
	StatusThrottle   int
	NormalizeHeaders func(f File) []string
	DataStartAt      int
	Config           *qor.Config
}

func New(res *Resource, cfg *qor.Config) *Exchange {
	return &Exchange{
		Resource:       res,
		JobThrottle:    1,
		DataStartAt:    1,
		StatusThrottle: 10,
		Config:         cfg,
		NormalizeHeaders: func(f File) (headers []string) {
			if f.TotalLines() <= 0 {
				return
			}

			return f.Line(0)
		},
	}
}

type ImportStatus struct {
	LineNum    int
	MetaValues *resource.MetaValues
	Errors     []error
}

type ImportInfo struct {
	TotalLines int
	Done       chan bool
	Error      chan error
}

type File interface {
	TotalLines() (num int)
	Line(l int) (fields []string)
}

func (ex *Exchange) Import(f File, log io.Writer) (err error) {
	doneChan := make(chan bool)
	errChan := make(chan error)
	importStatusChan := make(chan ImportStatus, ex.StatusThrottle)

	go ex.process(f, doneChan, errChan, importStatusChan, log)

	var statuses []ImportStatus
	// index := ex.DataStartAt
loop:
	for {
		select {
		case <-doneChan:
			break loop
		case err = <-errChan:
			break loop
		case ii := <-importStatusChan:
			statuses = append(statuses, ii)

			// if ii.LineNum != index {
			// 	continue
			// }

			// index, statuses = logInOrder(log, index, total, statuses)
		}
	}

	return
}

// func logInOrder(log io.Writer, index, total int, statuses []ImportStatus) (newIndex int, newStatuses []ImportStatus) {
// 	newIndex = index
// 	for _, status := range statuses {
// 		if status.LineNum != index {
// 			newStatuses = append(newStatuses, status)
// 			continue
// 		}

// 		newIndex += 1
// 		if len(status.Errors) == 0 {
// 			log.Write([]byte(fmt.Sprintf("%d/%d Done", index, total)))
// 		} else {
// 			log.Write([]byte(fmt.Sprintf("%d/%d %s", index, total, status.Errors)))
// 		}
// 	}

// 	if newIndex == index {
// 		return
// 	}

// 	if len(newStatuses) > 0 {
// 		return logInOrder(log, newIndex, total, newStatuses)
// 	}

// 	return index, newStatuses
// }

func (ex *Exchange) process(f File, doneChan chan bool, errChan chan error, importStatusChan chan ImportStatus, log io.Writer) {
	var wait sync.WaitGroup
	totalLines := f.TotalLines()
	wait.Add(totalLines - ex.DataStartAt)
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

	db := ex.Config.DB.Begin()
	res := ex.Resource
	headers := ex.NormalizeHeaders(f)
	for num := ex.DataStartAt; num < totalLines; num++ {
		throttle <- true
		if hasError && ex.StopOnError {
			goto rollback
		}

		go func(num int, importStatusChan chan ImportStatus) {
			importStatus := ImportStatus{LineNum: num}
			line := f.Line(num)
			defer func() {
				setError(len(importStatus.Errors) > 0)
				var msg string
				if len(importStatus.Errors) > 0 {
					for _, err := range importStatus.Errors {
						msg += err.Error() + "; "
					}
				} else {
					msg = abstractMsg(line)
				}
				log.Write([]byte(fmt.Sprintf("%d/%d: %s\n", num+1, totalLines, msg)))

				<-throttle
				importStatusChan <- importStatus
				wait.Done()
			}()

			vmap := map[string]string{}
			lineLen := len(line)
			for j, header := range headers {
				if j >= lineLen {
					break
				}

				vmap[header] = line[j]
			}

			importStatus.MetaValues, _ = res.getMetaValues(vmap, 0)
			processor := resource.DecodeToResource(res, res.NewStruct(), importStatus.MetaValues, nil)

			if err := processor.Initialize(); err != nil {
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
		errChan <- err
		return
	}
	doneChan <- true
	return

rollback:
	if err := db.Rollback().Error; err != nil {
		errChan <- err
	}
	errChan <- errors.New("exchange: encounter error in job processing")
	return
}

func abstractMsg(line []string) (msg string) {
	for i, field := range line {
		if i > 3 {
			return
		}
		msg += field + " "
	}

	return
}
