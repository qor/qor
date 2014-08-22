package worker

import (
	"bytes"
	"github.com/qor/qor/resource"

	"io"
	"strings"
)

type Adapter interface {
	Enqueue(metaValues *resource.MetaValues) Job
	Listen(worker Worker)
	GetProcessLog(Job) io.Reader
	LogWriter(Job) io.Writer
	Kill(Job)
}

type SampleAdapter struct {
}

func (SampleAdapter) Enqueue(metaValues *resource.MetaValues) Job {
	// push to job queue
	return Job{}
}

func (SampleAdapter) Listen(worker Worker) {
	// parse ARGV, to check it is in running job, if so run the job and exit program after finish

	// listen from job queue -> if there is a new job -> start a new process to run the job -> save process id
	// listen from job queue -> if there is a kill command -> kill the related process

	// the job queue handle schedule
}

func (SampleAdapter) GetProcessLog(job Job) io.Reader {
	return strings.NewReader("") // also need to a writer, it should read running job's logs
}

func (SampleAdapter) LogWriter(job Job) io.Writer {
	return bytes.NewBuffer([]byte{})
}

func (SampleAdapter) Kill(job Job) {
}
