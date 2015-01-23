package worker

type Queuer interface {
	Name() string
	Enqueue(job *QorJob) (err error)
	Purge(job *QorJob) (err error)
	Dequeue() (jobId uint64, err error)
}
