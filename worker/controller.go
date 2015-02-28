package worker

import (
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/qor/qor/admin"
	"github.com/qor/qor/utils"
)

var viewInject sync.Once

// TODO: UNDONE
func (w *Worker) InjectQorAdmin(a *admin.Admin) {
	viewInject.Do(func() {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/worker/views"))
		}
	})

	w.admin = a
	for _, job := range w.Jobs {
		job.initResource()
	}

	param := utils.ToParamString(w.Name)
	a.GetRouter().Get("/"+param+"/new", w.newJobPage)
	a.GetRouter().Post("/"+param+"/new", w.createJob)
	a.GetRouter().Get("/"+param+`/[\d]+`, w.showJob)
	a.GetRouter().Get("/"+param, w.indexPage)
}

func (w *Worker) indexPage(c *admin.Context) {
	var resource *admin.Resource
	for _, j := range w.Jobs {
		resource = j.Resource
	}
	qorJobs, err := c.SetResource(resource).FindAll()
	if err != nil {
		panic(err)
	}

	c.Execute("job/index", struct {
		*Worker
		QorJobs interface{}
	}{Worker: w, QorJobs: qorJobs})
}

func (w *Worker) newJobPage(c *admin.Context) {
	jobName := c.Request.URL.Query().Get("job")
	job, ok := w.Jobs[jobName]
	if !ok {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	c.SetResource(job.Resource)
	c.Execute("job/new", job)
}

// TODO: remove panics
func (w *Worker) createJob(c *admin.Context) {
	jobName := c.Request.URL.Query().Get("job")
	job, ok := w.Jobs[jobName]
	if !ok {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO: get current user
	qorJob := job.NewQorJob(0, time.Time{}, "tbd", DefaultJobCli)
	if errs := job.Resource.Decode(c, qorJob); len(errs) > 0 {
		panic(errs)
	}

	if err := jobDB.Save(qorJob).Error; err != nil {
		panic(err)
	}

	if err := job.Enqueue(qorJob); err != nil {
		panic(err)
	}

	http.Redirect(c.Writer, c.Request, qorJob.URL(), http.StatusSeeOther)
}

func (w *Worker) showJob(c *admin.Context) {
	parts := strings.Split(c.Request.URL.Path, "/")
	var job QorJob
	if err := jobDB.Where("id = " + parts[len(parts)-1]).Find(&job).Error; err != nil {
		panic(err)
	}

	c.Execute("job/show", &job)
}
