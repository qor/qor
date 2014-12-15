package worker

import "io"

type Queuer interface {
	Enqueue(job *Job) (err error)
	// Listen(worker *Worker)
	Dequeue() (job *Job, err error)
	// GetProcessLog(job *Job) io.Reader
	// LogWriter(job *Job) io.Writer
	// GetLogger(job *Job) io.ReadWriter

	// Kill(job *Job) error

	// SuspendJob(job *Job)
	// RerunJob(job *Job)
}

func RunJob(job *Job) {

}

func GetLogger(job *Job) (rw io.ReadWriter) {
	return
}

// type SampleAdapter struct{}

// // StartAt
// // Interval
// func (SampleAdapter) Enqueue(job *Job) {
// 	// push to job queue
// }

// func (SampleAdapter) Listen(worker *Worker) {
// 	// parse ARGV, to check it is in running job, if so run the job and exit program after finish

// 	// listen from job queue -> if there is a new job -> start a new process to run the job -> save process id
// 	// listen from job queue -> if there is a kill command -> kill the related process

// 	// the job queue handle schedule
// }

// func (SampleAdapter) GetProcessLog(job *Job) io.Reader {
// 	return strings.NewReader("") // also need to a writer, it should read running job's logs
// }

// func (SampleAdapter) LogWriter(job *Job) io.Writer {
// 	return bytes.NewBuffer([]byte{})
// }

// func (SampleAdapter) Kill(job *Job) bool {
// 	return false
// }
