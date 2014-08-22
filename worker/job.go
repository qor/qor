package worker

import (
	"github.com/qor/qor/resource"

	"io"
)

type Job struct {
	Id         string
	MetaValues *resource.MetaValues
	Worker     Worker
}

func (job Job) GetProcessLog() io.Reader {
	return job.Worker.GetProcessLog(job)
}

func (job Job) LogWriter() io.Writer {
	return job.Worker.LogWriter(job)
}

func (job Job) Kill() {
	job.Worker.Kill(job)
}
