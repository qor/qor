package worker

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/utils"
)

type Job struct {
	Name   string
	Queuer Queuer
	// Config *qor.Config

	Description string

	worker   *Worker
	Resource *admin.Resource
	metas    []*admin.Meta

	Handle    func(job *QorJob) error
	OnKill    func(job *QorJob) error
	OnStart   func(job *QorJob) error
	OnSuccess func(job *QorJob)
	OnFailed  func(job *QorJob)
}

// func (j *Job) ExtraInput(res *admin.Resource) {
// 	j.Resource = res
// }

func (j *Job) Run(job *QorJob) (err error) {
	if err = job.SavePID(); err != nil {
		return
	}
	logger, err := job.GetLogger()
	if err != nil {
		return
	}

	fmt.Fprintf(logger, "run job (%d) with pid (%d)\n", job.Id, job.PID)

	if j.OnStart != nil {
		if err = j.OnStart(job); err != nil {
			logger.Write([]byte("worker.onstart: " + err.Error() + "\n"))

			if err = job.UpdateStatus(StatusFailed); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err = job.UpdateStatus(StatusRunning); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		return
	}

	if err = j.Handle(job); err != nil {
		if err = job.UpdateStatus(StatusFailed); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if j.OnFailed != nil {
			j.OnFailed(job)
		}

		logger.Write([]byte("worker.hanlde: " + err.Error() + "\n"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if j.OnSuccess != nil {
		if err = job.UpdateStatus(StatusDone); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		j.OnSuccess(job)
	}

	if err = job.UpdateStatus(StatusDone); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(logger, "finish job (%d) with pid (%d)\n", job.Id, job.PID)

	return
}

var ErrJobDone = errors.New("job is finished")

func (j *Job) Kill(job *QorJob) (err error) {
	if j.OnKill != nil {
		if err = j.OnKill(job); err != nil {
			return
		}
	}

	switch job.Status {
	case StatusToRun:
		err = j.Queuer.Purge(job)
	case StatusRunning:
		if job.PID == 0 {
			return errors.New("pid is zero")
		}

		var process *os.Process
		process, err = os.FindProcess(job.PID)
		if err != nil {
			return
		}

		err = process.Kill()
	case StatusDone:
		return ErrJobDone
	}

	if err == nil {
		err = job.UpdateStatus(StatusKilled)
	}

	return
}

// func (j *Job) NewQorJob(interval uint64, startAt time.Time, by string) (job *QorJob, err error) {
// 	return j.NewQorJobWithCli(interval, startAt, by, DefaultJobCli)
// }

func (j *Job) NewQorJob(interval uint64, startAt time.Time, by, cli string) (job *QorJob) {
	job = &QorJob{
		Interval:   interval,
		StartAt:    startAt,
		JobName:    j.Name,
		WorkerName: j.worker.Name,
		Cli:        cli,
		Status:     StatusToRun,
		By:         by,
		// ExtraInputs: extraInputs,
	}

	// if err = jobDB.Save(job).Error; err != nil {
	// 	return
	// }

	// if err = j.Queuer.Enqueue(job); err != nil {
	// 	return
	// }

	// if job.QueueJobId != "" {
	// 	if err = jobDB.Save(job).Error; err != nil {
	// 		return
	// 	}
	// }

	return
}

func (j *Job) Enqueue(job *QorJob) (err error) {
	if err = j.Queuer.Enqueue(job); err != nil {
		return
	}

	if job.QueueJobId != "" {
		if err = jobDB.Save(job).Error; err != nil {
			return
		}
	}

	return
}

func (j *Job) initResource() {
	qorjob := j.worker.admin.AddResource(&QorJob{}, &admin.Config{Name: j.Name + "-QorJob", Invisible: true})
	// qorjob.IndexAttrs("Id", "QueueJobId", "Interval", "StartAt", "Cli", "WorkerName", "Status", "PID", "RunCounter", "FailCounter", "SuccessCounter", "KillCounter")
	qorjob.NewAttrs("Interval", "StartAt")

	scopes := map[string]string{
		"running": StatusRunning,
		"done":    StatusDone,
		"failed":  StatusFailed,
	}
	for n, s := range scopes {
		qorjob.Scope(&admin.Scope{
			Name: n,
			Handle: func(db *gorm.DB, context *qor.Context) *gorm.DB {
				return db.Where("status = ?", s)
			},
		})
	}

	// qorjob.Meta(&admin.Meta{Name: "WorkerName", Type: "select_one", Collection: func(interface{}, *qor.Context) [][]string {
	// 	var keys [][]string
	// 	for k, _ := range w.jobs {
	// 		keys = append(keys, []string{k, k})
	// 	}
	// 	return keys
	// }})

	j.Resource = qorjob
}

func (j *Job) Meta(meta *admin.Meta) {
	if meta.Valuer == nil {
		meta.Valuer = func(val interface{}, ctx *qor.Context) interface{} {
			ev := val.(*QorJob).ExtraValue
			if ev == nil {
				return ""
			}
			return (*ev)[meta.Name]
		}
	}
	if meta.Setter == nil {
		meta.Setter = func(val interface{}, metaValues *resource.MetaValues, ctx *qor.Context) {
			mv := metaValues.Get(meta.Name)
			if mv == nil {
				return
			}
			q, ok := val.(*QorJob)
			if !ok {
				return
			}

			if meta.Type != "file" {
				val, ok := mv.Value.([]string)
				if !ok || len(val) == 0 {
					return
				}

				if q.ExtraValue == nil {
					q.ExtraValue = &ExtraInput{}
				}
				ev := *(q.ExtraValue)
				ev[mv.Name] = val[0]
				return
			}

			headers, ok := mv.Value.([]*multipart.FileHeader)
			if !ok || len(headers) == 0 {
				return
			}
			h := headers[0]
			name := fmt.Sprintf("%s-%d-%s", strings.Replace(j.Name, "/", "-", -1), time.Now().UnixNano(), h.Filename)
			path := filepath.Join(WorkerDataPath, name)
			dst, err := os.Create(path)
			if err != nil {
				fmt.Printf("worker: os.Create(%s): %s\n", path, err)
				return
			}
			src, err := h.Open()
			if err != nil {
				fmt.Println("worker: h.Open():", err)
				return
			}
			if _, err := io.Copy(dst, src); err != nil {
				fmt.Println("worker: io.Copy(dst, src):", err)
				return
			}
			if q.ExtraFile == nil {
				q.ExtraFile = &ExtraInput{}
			}
			ef := *(q.ExtraFile)
			ef[mv.Name] = name
		}
	}
	j.Resource.Meta(meta)
	j.metas = append(j.metas, meta)

	attrs := []string{"Interval", "StartAt"}
	for _, meta := range j.metas {
		attrs = append(attrs, meta.GetName())
	}
	j.Resource.NewAttrs(attrs...)
}

func (j *Job) URL() string {
	return fmt.Sprintf("%s/%s/new?job=%s", j.worker.admin.GetRouter().Prefix, utils.ToParamString(j.worker.Name), j.Name)
}
