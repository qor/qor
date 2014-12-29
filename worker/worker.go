package worker

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/qor/qor/resource"

	"github.com/qor/qor/roles"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

var (
	jobDB         *gorm.DB
	jobId         uint64
	workers       = map[string]*Worker{}
	queuers       = map[string]Queuer{}
	DefaultJobCli = strings.Join(os.Args, " ")
)

func init() {
	flag.Uint64Var(&jobId, "job-id", 0, "qor job id")
}

// SetJobDB will run a auto migration for creating table jobs
func SetJobDB(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&Job{}).Error
	if err != nil {
		return
	}

	jobDB = db

	return
}

// TODO: UNDONE
func SetAdmin(a *admin.Admin) {
	job := a.NewResource(&Job{})
	job.IndexAttrs("Id", "QueueJobId", "Interval", "StartAt", "Cli", "WorkerName", "Status", "PID", "RunCounter", "FailCounter", "SuccessCounter", "KillCounter")
	job.NewAttrs("Interval", "StartAt", "WorkerName")

	job.Meta(&resource.Meta{Name: "WorkerName", Type: "select_one", Collection: func(interface{}, *qor.Context) [][]string {
		var keys [][]string
		for k, _ := range workers {
			keys = append(keys, []string{k, k})
		}
		return keys
	}})

	admin.RegisterViewPath(os.Getenv("GOPATH") + "/src/github.com/qor/qor/worker/templates")
	a.GetRouter().Get("/job/new", newJobPage)
	a.GetRouter().Get("/job/switch_worker", switchWorker)
}

func newJobPage(c *admin.Context) {
	var res *admin.Resource
	for _, w := range workers {
		res = w.resource
		break
	}
	content := admin.Content{Context: c, Admin: c.Admin, Resource: res, Action: "new"}
	c.Admin.Render("new", content, roles.Create)
}

func switchWorker(c *admin.Context) {
	wname := c.Request.FormValue("name")
	w, ok := workers[wname]
	if !ok {
		c.Writer.WriteHeader(http.StatusBadRequest)
		c.Writer.Write([]byte("worker does not exist"))
		return
	}

	content := admin.Content{Context: c, Admin: c.Admin, Resource: w.resource, Action: "new"}
	c.Admin.Render("worker", content, roles.Create)
}

// Listen will parse an flag named as "job-id". If the job-id is zero, it
// will run as queue listen server. Otherwise, it will run a specific job
// and terminate the process after the job is run.
//
// It must be executed before http.ListenAndServer
func Listen() {
	flag.Parse()

	if jobId > 0 {
		var job Job
		if err := jobDB.Where("id = ?", jobId).Find(&job).Error; err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		var logger io.Writer
		var err error
		if logger, err = job.GetLogger(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		var w *Worker
		if w, err = job.GetWorker(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		if err = w.Run(&job); err != nil {
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

func New(queuer Queuer, name string, handle func(job *Job) error) (w *Worker) {
	w = &Worker{
		Name:   name,
		Handle: handle,
		Queuer: queuer,
	}

	workers[w.Name] = w
	queuers[queuer.Name()] = queuer

	return
}

type Worker struct {
	Name   string
	Queuer Queuer
	Config *qor.Config

	resource *admin.Resource

	Handle    func(job *Job) error
	OnKill    func(job *Job) error
	OnStart   func(job *Job) error
	OnSuccess func(job *Job)
	OnFailed  func(job *Job)
}

func (w *Worker) ExtraInput(res *admin.Resource) {
	w.resource = res
}

func (w *Worker) Run(job *Job) (err error) {
	if err = job.SavePID(); err != nil {
		return
	}
	logger, err := job.GetLogger()
	if err != nil {
		return
	}

	fmt.Fprintf(logger, "run job (%d) with pid (%d)\n", job.Id, job.PID)

	if w.OnStart != nil {
		if err = w.OnStart(job); err != nil {
			logger.Write([]byte("worker.onstart: " + err.Error() + "\n"))

			if err = job.UpdateStatus(JobFailed); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	if err = job.UpdateStatus(JobRunning); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		return
	}

	if err = w.Handle(job); err != nil {
		if err = job.UpdateStatus(JobFailed); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		if w.OnFailed != nil {
			w.OnFailed(job)
		}

		logger.Write([]byte("worker.hanlde: " + err.Error() + "\n"))
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if w.OnSuccess != nil {
		if err = job.UpdateStatus(JobRun); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		w.OnSuccess(job)
	}

	if err = job.UpdateStatus(JobRun); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(logger, "finish job (%d) with pid (%d)\n", job.Id, job.PID)

	return
}

var ErrJobRun = errors.New("job is already run")

func (w *Worker) Kill(job *Job) (err error) {
	if w.OnKill != nil {
		if err = w.OnKill(job); err != nil {
			return
		}
	}

	switch job.Status {
	case JobToRun:
		err = w.Queuer.Purge(job)
	case JobRunning:
		if job.PID == 0 {
			return errors.New("pid is zero")
		}

		var process *os.Process
		process, err = os.FindProcess(job.PID)
		if err != nil {
			return
		}

		err = process.Kill()
	case JobRun:
		return ErrJobRun
	}

	if err == nil {
		err = job.UpdateStatus(JobKilled)
	}

	return
}

func (w *Worker) NewJob(interval uint64, startAt time.Time) (job *Job, err error) {
	return w.NewJobWithCli(interval, startAt, DefaultJobCli)
}

func (w *Worker) NewJobWithCli(interval uint64, startAt time.Time, cli string) (job *Job, err error) {
	job = &Job{
		Interval:   interval,
		StartAt:    startAt,
		WorkerName: w.Name,
		Cli:        cli,
	}
	if err = jobDB.Save(job).Error; err != nil {
		return
	}

	err = w.Queuer.Enqueue(job)

	if job.QueueJobId != "" {
		if err = jobDB.Save(job).Error; err != nil {
			return
		}
	}

	return
}
