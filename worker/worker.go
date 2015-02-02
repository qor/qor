package worker

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
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

// SetJobDB will run a auto migration for creating table jobs
func SetJobDB(db *gorm.DB) (err error) {
	err = db.AutoMigrate(&QorJob{}).Error
	if err != nil {
		return
	}

	jobDB = db

	return
}

type Worker struct {
	Name string
	jobs map[string]*Job
}

func New(name string) *Worker {
	w := &Worker{Name: name, jobs: map[string]*Job{}}
	workers[name] = w
	return w
}

var viewInject sync.Once

// TODO: UNDONE
func (w *Worker) InjectQorAdmin(a *admin.Admin) {
	// job := a.NewResource(&QorJob{})
	// job.IndexAttrs("Id", "QueueJobId", "Interval", "StartAt", "Cli", "WorkerName", "Status", "PID", "RunCounter", "FailCounter", "SuccessCounter", "KillCounter")
	// job.NewAttrs("Interval", "StartAt", "WorkerName")

	// job.Meta(&resource.Meta{Name: "WorkerName", Type: "select_one", Collection: func(interface{}, *qor.Context) [][]string {
	// 	var keys [][]string
	// 	for k, _ := range w.jobs {
	// 		keys = append(keys, []string{k, k})
	// 	}
	// 	return keys
	// }})

	viewInject.Do(func() {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/worker/views"))
		}
	})

	a.GetRouter().Get("/"+w.Name, w.indexPage)
	a.GetRouter().Get("/"+w.Name+"/job/new", w.newJobPage)
	a.GetRouter().Get("/"+w.Name+"/job/switch_worker", w.switchWorker)
}

func (w *Worker) AllJobs() (jobs []string) {
	for k, _ := range w.jobs {
		jobs = append(jobs, k)
	}

	return
}

func (w *Worker) indexPage(c *admin.Context) {
	var qorJobs []QorJob
	if err := jobDB.Where("worker_name = ?", w.Name).Find(&qorJobs).Error; err != nil {
		// c.Admin.RenderError(err, http.StatusInternalServerError, c)
		return
	}

	c.Execute("job/new", struct {
		Jobs    []string
		QorJobs []QorJob
	}{Jobs: w.AllJobs(), QorJobs: qorJobs})
}

func (w *Worker) newJobPage(c *admin.Context) {
	var res *admin.Resource
	for _, j := range w.jobs {
		res = j.resource
		break
	}
	// content := admin.Content{Context: c, Admin: c.Admin, Resource: res, Action: "new"}
	// c.Admin.Render("new", content, roles.Create)
	c.Execute("new", res)
}

func (w *Worker) switchWorker(c *admin.Context) {
	// wname := c.Request.FormValue("name")
	// w, ok := workers[wname]
	// if !ok {
	// 	c.Writer.WriteHeader(http.StatusBadRequest)
	// 	c.Writer.Write([]byte("worker does not exist"))
	// 	return
	// }

	// content := admin.Content{Context: c, Admin: c.Admin, Resource: w.resource, Action: "new"}
	// c.Admin.Render("worker", content, roles.Create)
}

// Listen will parse an flag named as "job-id". If the job-id is zero, it
// will run as queue listen server. Otherwise, it will run a specific job
// and terminate the process after the job is run.
//
// It must be executed before http.ListenAndServer
func Listen() {
	flag.Parse()

	if jobId > 0 {
		var job QorJob
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

		var j *Job
		if j, err = job.GetWorker(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(logger, err)
			os.Exit(1)
		}

		if err = j.Run(&job); err != nil {
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

func (w Worker) NewJob(queuer Queuer, name string, handle func(job *QorJob) error) (j *Job) {
	j = &Job{
		Name:   name,
		Handle: handle,
		Queuer: queuer,
	}

	w.jobs[j.Name] = j
	queuers[queuer.Name()] = queuer

	return
}

type Job struct {
	Name   string
	Queuer Queuer
	Config *qor.Config

	resource *admin.Resource

	Handle    func(job *QorJob) error
	OnKill    func(job *QorJob) error
	OnStart   func(job *QorJob) error
	OnSuccess func(job *QorJob)
	OnFailed  func(job *QorJob)
}

func (j *Job) ExtraInput(res *admin.Resource) {
	j.resource = res
}

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

	if err = j.Handle(job); err != nil {
		if err = job.UpdateStatus(JobFailed); err != nil {
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
		if err = job.UpdateStatus(JobRun); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		j.OnSuccess(job)
	}

	if err = job.UpdateStatus(JobRun); err != nil {
		fmt.Fprintf(logger, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(logger, "finish job (%d) with pid (%d)\n", job.Id, job.PID)

	return
}

var ErrJobRun = errors.New("job is already run")

func (j *Job) Kill(job *QorJob) (err error) {
	if j.OnKill != nil {
		if err = j.OnKill(job); err != nil {
			return
		}
	}

	switch job.Status {
	case JobToRun:
		err = j.Queuer.Purge(job)
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

func (j *Job) NewJob(interval uint64, startAt time.Time) (job *QorJob, err error) {
	return j.NewJobWithCli(interval, startAt, DefaultJobCli)
}

func (j *Job) NewJobWithCli(interval uint64, startAt time.Time, cli string) (job *QorJob, err error) {
	job = &QorJob{
		Interval:   interval,
		StartAt:    startAt,
		WorkerName: j.Name,
		Cli:        cli,
	}
	if err = jobDB.Save(job).Error; err != nil {
		return
	}

	err = j.Queuer.Enqueue(job)

	if job.QueueJobId != "" {
		if err = jobDB.Save(job).Error; err != nil {
			return
		}
	}

	return
}
