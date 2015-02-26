package worker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

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
	var qorJobs []QorJob
	if err := jobDB.Where("worker_name = ?", w.Name).Order("id desc").Find(&qorJobs).Error; err != nil {
		panic(err)
	}

	c.Execute("job/index", struct {
		*Worker
		QorJobs []QorJob
	}{Worker: w, QorJobs: qorJobs})
}

func (w *Worker) newJobPage(c *admin.Context) {
	// var res *admin.Resource
	jobName := c.Request.URL.Query().Get("job")
	job, ok := w.Jobs[jobName]
	if !ok {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	c.SetResource(job.Resource)

	// content := admin.Content{Context: c, Admin: c.Admin, Resource: res, Action: "new"}
	// c.Admin.Render("new", content, roles.Create)
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

	var metaors []resource.Metaor
	for _, m := range job.Resource.NewMetas() {
		metaors = append(metaors, m)
	}
	mvs, err := resource.ConvertFormToMetaValues(c.Request, metaors, "QorResource.")
	if err != nil {
		panic(err)
	}

	interval, err := strconv.ParseUint(mvs.Get("Interval").Value.([]string)[0], 10, 64)
	if err != nil {
		panic(err)
	}
	startAt, err := time.Parse("2006-01-02T15:04", mvs.Get("StartAt").Value.([]string)[0])
	if err != nil {
		panic(err)
	}
	inputs, err := marshalMetaValues(mvs)
	if err != nil {
		panic(err)
	}

	// TODO: support custom JobCli
	qorjob, err := job.NewQorJob(interval, startAt, inputs)
	if err != nil {
		panic(err)
	}

	// TODO: turn this into a method?
	url := fmt.Sprintf("/%s/%s/%d", w.admin.GetRouter().Prefix, utils.ToParamString(w.Name), qorjob.Id)
	http.Redirect(c.Writer, c.Request, url, http.StatusSeeOther)
}

func marshalMetaValues(mvs *resource.MetaValues) (string, error) {
	m := map[string]interface{}{}
	for _, mv := range mvs.Values {
		if mv.Name == "Interval" || mv.Name == "StartAt" {
			continue
		}

		m[mv.Name] = mv.Value
	}
	r, err := json.Marshal(m)
	return string(r), err
}

func (w *Worker) showJob(c *admin.Context) {
	parts := strings.Split(c.Request.URL.Path, "/")
	var job QorJob
	if err := jobDB.Where("id = " + parts[len(parts)-1]).Find(&job).Error; err != nil {
		panic(err)
	}

	c.Execute("job/show", &job)
}
