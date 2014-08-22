package worker

import (
	"github.com/qor/qor/resource"

	"io"
	"strings"
)

type Adapter interface {
	Enqueue(metaValues *resource.MetaValues) (jobId string)
	Listen()
	GetProcessLog(jobId string) io.Reader
}

type SampleAdapter struct {
}

func (SampleAdapter) Enqueue(metaValues *resource.MetaValues) (jobId string) {
	// push to job queue
	return ""
}

func (SampleAdapter) Listen() {
	// parse ARGV, to check it is in running job, if so run the job and exit program after finish

	// listen from job queue -> if there is a new job -> start a new process to run the job -> save process id
	// listen from job queue -> if there is a kill command -> kill the related process

	// the job queue handle schedule
}

func (SampleAdapter) GetProcessLog(jobId string) io.Reader {
	return strings.NewReader("") // also need to a writer, it should read running job's logs
}
