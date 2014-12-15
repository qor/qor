package worker

import "time"

type Job struct {
	Id uint64

	// unit: minute
	// 0 to run job only once
	Interval int64

	// zero time value to execute job immediately
	StartAt time.Time

	WokerSetName string
	WorkerName   string
	Worker       Worker
	// Errors []error

	// MetaValues *resource.MetaValues
}

// func (job *Job) AddErr(err error) {
// 	job.Errors = append(job.Errors, err)
// }

// func (job *Job) GetProcessLog() io.Reader {
// 	return job.Worker.GetProcessLog(job)
// }

// func (job *Job) LogWriter() io.Writer {
// 	return job.Worker.LogWriter(job)
// }

// func (job *Job) Kill() {
// 	if job.Worker.Kill(job) {
// 		job.Worker.OnKill(job)
// 	}
// }
