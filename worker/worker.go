package worker

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/admin"
)

var (
	jobDB *gorm.DB
	jobId uint64
	// workers       = map[string]*Job{}
	workers       = map[string]*Worker{}
	queuers       = map[string]Queuer{}
	DefaultJobCli = strings.Join(os.Args, " ")
)

func init() {
	flag.Uint64Var(&jobId, "job-id", 0, "qor job id")
}

func SetJobDB(db *gorm.DB) {
	jobDB = db
}

// Listen will parse an flag named as "job-id". If the job-id is zero, it
// will run as queue listen server. Otherwise, it will run a specific job
// and terminate the process after the job is run.
//
// It must be executed before http.ListenAndServer
func Listen() {
	flag.Parse()

	if jobId > 0 {
		var qorjob QorJob
		if err := jobDB.Where("id = ?", jobId).Find(&qorjob).Error; err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var logger io.Writer
		var err error
		if logger, err = qorjob.GetLogger(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		var job *Job
		if job = qorjob.GetJob(); job == nil {
			fmt.Fprintln(os.Stderr, "job not found")
			fmt.Fprintln(logger, "job not found")
			os.Exit(1)
		}

		if err = job.Run(&qorjob); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		os.Exit(0)
	} else {
		for _, queuer := range queuers {
			go func() {
				for {
					jobId, err := queuer.Dequeue()
					fmt.Println("dequeue job", jobId)
					if err != nil {
						fmt.Println("qor.worker.dequeue.error:", err)
					} else {
						go RunJob(jobId)
					}
				}
			}()
		}
	}
}

type Worker struct {
	Name  string
	admin *admin.Admin
	Jobs  map[string]*Job
}

func New(name string) *Worker {
	w := &Worker{Name: name, Jobs: map[string]*Job{}}
	workers[name] = w
	return w
}

func (w *Worker) NewJob(queuer Queuer, name, desc string, handle func(job *QorJob) error) (j *Job) {
	j = &Job{
		Name:        name,
		Handle:      handle,
		Queuer:      queuer,
		Description: desc,
		worker:      w,
	}

	if w.admin != nil {
		j.initResource()
	}

	w.Jobs[j.Name] = j
	queuers[queuer.Name()] = queuer

	return
}

func (w *Worker) ResourceName() string {
	return w.Name
}

// func (w *Worker) ResourceParam() string {
// 	return strings.ToLower(strings.Replace(w.Name, " ", "_", -1))
// }
