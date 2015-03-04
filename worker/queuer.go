package worker

type Queuer interface {
	// Name returns the Queuer's identifier
	Name() string
	// Enqueue pushes a job to a queue, also a Queuer could set a id value (string)
	// in QorJob's QueueJobId if needed
	// Interval
	// StartAt
	Enqueue(job *QorJob) (err error)
	// Purge removes a job from a queue
	Purge(job *QorJob) (err error)
	// Dequeue blocks the process until a job id (and error if any) is returned
	Dequeue() (jobId uint64, err error)
}
