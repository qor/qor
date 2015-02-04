package worker

import (
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
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
	for _, job := range w.jobs {
		job.initResource()
	}

	param := utils.ToParamString(w.Name)
	a.GetRouter().Get("/"+param+"/new", w.newJobPage)
	a.GetRouter().Post("/"+param+"/new", w.createJob)
	a.GetRouter().Get("/"+param, w.indexPage)
	a.GetRouter().Get("/"+param+"/switch_worker", w.switchWorker)
}

func (w *Worker) AllJobs() (jobs []string) {
	for k, _ := range w.jobs {
		jobs = append(jobs, k)
	}

	return
}

func (w *Worker) indexPage(c *admin.Context) {
	var qorJobs []QorJob
	if err := jobDB.Where("worker_name = ?", w.Name).Order("id desc").Find(&qorJobs).Error; err != nil {
		// c.Admin.RenderError(err, http.StatusInternalServerError, c)
		return
	}

	c.Execute("job/index", struct {
		Jobs    []string
		QorJobs []QorJob
	}{Jobs: w.AllJobs(), QorJobs: qorJobs})
}

func (w *Worker) newJobPage(c *admin.Context) {
	// var res *admin.Resource
	jobName := c.Request.URL.Query().Get("job")
	job, ok := w.jobs[jobName]
	if !ok {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	c.SetResource(job.Resource)

	// content := admin.Content{Context: c, Admin: c.Admin, Resource: res, Action: "new"}
	// c.Admin.Render("new", content, roles.Create)
	c.Execute("job/new", job)
}

func (w *Worker) createJob(c *admin.Context) {
	jobName := c.Request.URL.Query().Get("job")
	job, ok := w.jobs[jobName]
	if !ok {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	var metaors []resource.Metaor
	for _, m := range job.Resource.NewMetas() {
		metaors = append(metaors, m)
	}
	c.Request.ParseForm()
	mvs, err := resource.ConvertFormToMetaValues(c.Request, metaors, "QorResource.")
	if err != nil {
		panic(err)
	}

	// fmt.Printf("--> %+v\n", mvs.Get("StartAt").Value)
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
