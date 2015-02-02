package worker

import (
	"github.com/qor/qor/utils"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/qor/qor/admin"
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
		job.initResource(w)
	}

	param := utils.ToParamString(w.Name)
	a.GetRouter().Get("/"+param, w.indexPage)
	a.GetRouter().Get("/"+param+"/new", w.newJobPage)
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
