package exchange

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

// TODO: Formated Value
//  	def formatted_description=(description)
//  		self.description = description.to_s.gsub(/[\r\n]/, "")
//  	end
//
//  	def formatted_description
//  		self.description.to_s.gsub(/[\r\n]/, "")
//  	end

type Exchange struct {
	Resource         *Resource
	StopOnError      bool
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

type logger struct {
	log    io.Writer
	locker sync.Mutex
}

func (c *logger) Write(data []byte) (int, error) {
	c.locker.Lock()
	defer c.locker.Unlock()
	return c.log.Write(data)
}

func (ex *Exchange) Import(f File, log io.Writer, ctx *qor.Context) (err error) {
	doneChan := make(chan bool)
	errChan := make(chan error)
	importStatusChan := make(chan ImportStatus, ex.StatusThrottle)

	go ex.process(f, doneChan, errChan, importStatusChan, &logger{log: log}, ctx)

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

func (ex *Exchange) process(f File, doneChan chan bool, errChan chan error, importStatusChan chan ImportStatus, log io.Writer, ctx *qor.Context) {
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

	db := ctx.DB().Begin()
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
					msg = digestMsg(line)
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
			processor := resource.DecodeToResource(res, res.NewStruct(), importStatus.MetaValues, ctx)

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

func digestMsg(line []string) (msg string) {
	for i, field := range line {
		if i > 3 {
			return
		}
		msg += field + " "
	}

	return
}

// Export will format data into csv string and write it into a writer, by the definitions of metas.
// Note: records must be a slice or Export will panic.
func (ex *Exchange) Export(records interface{}, w io.Writer, ctx *qor.Context) (err error) {
	var headers []string
	walkMetas(ex.Resource, ctx, nil, func(_ resource.Resourcer, metaor resource.Metaor, _ interface{}) {
		if meta := metaor.GetMeta(); meta.Resource == nil {
			headers = append(headers, meta.Label)
		}
	})

	var fieldMaps []map[string]string
	var walker func(resource.Resourcer, resource.Metaor, interface{})
	fieldSizes := map[string]int{}
	recordsx := reflect.ValueOf(records)
	for i, count := 0, recordsx.Len(); i < count; i++ {
		record := recordsx.Index(i).Interface()
		fieldMap := map[string]string{}
		labelCounter := map[string]int{}
		walker = func(res resource.Resourcer, metaor resource.Metaor, record interface{}) {
			// fmt.Printf("--> %+v\n", record)
			if meta := metaor.GetMeta(); meta.Resource == nil {
				metaRes, ok := res.(*Resource)
				if !ok {
					return
				}

				value := fmt.Sprintf("%v", meta.Value(record, ctx))
				label := meta.Label
				labelCounter[label] = labelCounter[label] + 1
				if metaRes.HasSequentialColumns {
					index := labelCounter[label]
					if size, ok := fieldSizes[meta.Label]; !ok {
						fieldSizes[meta.Label] = index
					} else if size < index {
						fieldSizes[meta.Label] = index
					}

					label = fmt.Sprintf("%s %#02d", label, index)
				} else if metaRes.MultiDelimiter != "" {
					prev := fieldMap[label]
					if prev != "" {
						prev += metaRes.MultiDelimiter
					}
					value = prev + value
				}

				fieldMap[label] = value
			} else if fieldValue := reflect.ValueOf(record); fieldValue.Kind() == reflect.Slice {
				for j, count := 0, fieldValue.Len(); j < count; j++ {
					metaRecord := fieldValue.Index(j).Interface()
					walkMetas(meta.Resource, ctx, metaRecord, walker)
				}
			}
		}

		walkMetas(ex.Resource, ctx, record, walker)
		fieldMaps = append(fieldMaps, fieldMap)
	}

	headers = populateHeaders(headers, fieldSizes)
	w.Write([]byte(strings.Join(headers, ",") + "\n"))

	for _, fieldMap := range fieldMaps {
		var fields []string
		for _, header := range headers {
			field := fieldMap[header]
			if strings.Contains(field, ",") {
				field = "\"" + field + "\""
			}
			fields = append(fields, field)
		}
		w.Write([]byte(strings.Join(fields, ",") + "\n"))
	}

	return
}

// populateHeaders append index to headers for meta with HasSequentialColumns.
//   ["name", "age", "address"] + {name: 2, address: 1} => ["name 01", "name 02", "age", "address 01"]
// TODO: should take empty value into consideration
func populateHeaders(headers []string, fieldSizes map[string]int) (newHeaders []string) {
	for _, header := range headers {
		if size, ok := fieldSizes[header]; ok {
			for i := 0; i < size; i++ {
				if strings.Contains(header, ",") {
					newHeaders = append(newHeaders, fmt.Sprintf("\"%s %#02d\"", header, i+1))
				} else {
					newHeaders = append(newHeaders, fmt.Sprintf("%s %#02d", header, i+1))
				}
			}
		} else {
			if strings.Contains(header, ",") {
				newHeaders = append(newHeaders, "\""+header+"\"")
			} else {
				newHeaders = append(newHeaders, header)
			}
		}
	}

	return
}

// TODO: support pointer type
func walkMetas(resx resource.Resourcer, ctx *qor.Context, record interface{}, walker func(resource.Resourcer, resource.Metaor, interface{})) {
	res, ok := resx.(*Resource)
	if !ok {
		return
	}
	for _, header := range res.HeadersInOrder {
		metaor := res.Metas[header]
		walker(resx, metaor, record)
		if resx := metaor.GetMeta().Resource; resx != nil {
			if record != nil {
				metaRecord := deepIndirect(reflect.ValueOf(metaor.GetMeta().Value(record, ctx)))
				switch metaRecord.Kind() {
				case reflect.Struct:
					walkMetas(resx, ctx, metaRecord.Interface(), walker)
				case reflect.Slice:
					walker(resx, metaor, metaRecord.Interface())
				}
			} else {
				walkMetas(resx, ctx, nil, walker)
			}
		}
	}
}

func deepIndirect(val reflect.Value) reflect.Value {
	if val.Kind() == reflect.Ptr {
		return deepIndirect(val.Elem())
	}

	return val
}
