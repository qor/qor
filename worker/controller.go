package worker

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/qor/qor/admin"
)

var viewInject sync.Once

// TODO: UNDONE
func (w *Worker) InjectQorAdmin(res *admin.Resource) {
	viewInject.Do(func() {
		for _, gopath := range strings.Split(os.Getenv("GOPATH"), ":") {
			admin.RegisterViewPath(path.Join(gopath, "src/github.com/qor/qor/worker/views"))
		}
	})

	w.admin = res.GetAdmin()
	for _, job := range w.Jobs {
		job.initResource()
	}

	router := res.GetAdmin().GetRouter()
	router.Get("/"+res.ToParam()+"/new", wrap(w.newJobPage))
	router.Post("/"+res.ToParam()+"/new", wrap(w.createJob))
	router.Post("^/"+res.ToParam()+`/[\d]+/kill$`, wrap(w.killJob))
	router.Get("/"+res.ToParam()+`/[\d]+$`, wrap(w.showJob))
	router.Post("/"+res.ToParam()+`/[\d]+/stop`, wrap(w.stopJob))
	router.Get("/"+res.ToParam(), wrap(w.indexPage))
}

func wrap(h admin.Handle) admin.Handle {
	return func(c *admin.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}
			var err error
			if er, ok := r.(error); !ok {
				err = fmt.Errorf("%v", r)
			} else {
				err = er
			}

			log.Println(err)
			debug.PrintStack()

			c.Writer.WriteHeader(http.StatusInternalServerError)
			c.Writer.Write([]byte(err.Error()))
		}()

		h(c)
	}
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

func (w *Worker) stopJob(c *admin.Context) {
	parts := strings.Split(c.Request.URL.Path, "/")
	var qj QorJob
	if err := jobDB.Where("id = ?", parts[len(parts)-2]).First(&qj).Error; err != nil {
		panic(err)
	}
	qj.Stopped = true
	if err := jobDB.Save(&qj).Error; err != nil {
		panic(err)
	}
	if job := qj.GetJob(); job != nil {
		if err := job.Queuer.Purge(&qj); err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("job %q not exist", qj.JobName))
	}

	http.Redirect(c.Writer, c.Request, strings.Replace(c.Request.URL.Path, "/stop", "", 1), http.StatusSeeOther)
}

func (w *Worker) killJob(c *admin.Context) {
	parts := strings.Split(c.Request.URL.Path, "/")
	var qj QorJob
	if err := jobDB.Where("id = ?", parts[len(parts)-2]).First(&qj).Error; err != nil {
		panic(err)
	}
	if err := qj.GetJob().Kill(&qj); err != nil {
		panic(err)
	}
	http.Redirect(c.Writer, c.Request, c.Request.Referer(), http.StatusSeeOther)
}
