package worker

import (
	"github.com/qor/qor/resource"

	"io"
)

type Job struct {
	Id         string
	MetaValues *resource.MetaValues
	Worker     Worker
	Errors     []error
}

func (job *Job) AddErr(err error) {
	job.Errors = append(job.Errors, err)
}

func (job *Job) GetProcessLog() io.Reader {
	return job.Worker.GetProcessLog(job)
}

func (job *Job) LogWriter() io.Writer {
	return job.Worker.LogWriter(job)
}

func (job *Job) Kill() {
	if job.Worker.Kill(job) {
		job.Worker.OnKill(job)
	}
}
