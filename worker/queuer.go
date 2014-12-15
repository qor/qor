package worker

type Queuer interface {
	Enqueue(job *Job) (err error)
	Purge(job *Job) (err error)
	Dequeue() (jobId uint64, err error)
}
