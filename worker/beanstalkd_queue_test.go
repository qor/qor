package worker

import (
	"strconv"
	"testing"
	"time"

	"github.com/manveru/gostalk/gostalkc"
)

func TestBeanstalkdQueue(t *testing.T) {
	queue := NewBeanstalkdQueue("beanstalkd-test", "localhost:11300")
	client, err := queue.newClient()
	if err != nil {
		t.Fatal(err)
	}
	{
		job := &QorJob{
			ID:       1,
			Interval: 0,
			StartAt:  time.Now(),
		}

		err := queue.Enqueue(job)
		if err != nil {
			t.Error("enqueue error: ", err)
		}
		if job.QueueJobId == "" {
			t.Error("QueueJobId should not be empty")
		}
		jobId, err := queue.Dequeue()
		if err != nil {
			t.Error("dequeue error: ", err)
		}
		if jobId != 1 {
			t.Error("jobID: expect %d got %d", 1, jobId)
		}
	}
	{
		job := &QorJob{
			ID:       2,
			Interval: 0,
			StartAt:  time.Now().Add(time.Second * 5),
		}

		err := queue.Enqueue(job)
		if err != nil {
			t.Error("enqueue error: ", err)
		}
		if job.QueueJobId == "" {
			t.Error("QueueJobId should not be empty")
		}
		jobId, err := queue.Dequeue()
		if err != nil {
			t.Error("dequeue error: ", err)
		}
		if jobId != 2 {
			t.Error("jobID: expect %d got %d", 2, jobId)
		}
	}
	{
		parseInterval = func(interval uint64) string {
			return "2"
		}
		job := &QorJob{
			ID:       3,
			Interval: 2,
			StartAt:  time.Now(),
		}

		err := queue.Enqueue(job)
		if err != nil {
			t.Error("enqueue error: ", err)
		}
		if job.QueueJobId == "" {
			t.Error("QueueJobId should not be empty")
		}
		for i := 0; i < 3; i++ {
			jobId, err := queue.Dequeue()
			if err != nil {
				t.Error("dequeue error: ", err)
			}
			if jobId != 3 {
				t.Error("jobID: expect %d got %d", 3, jobId)
			}
		}

		err = queue.Purge(job)
		if err != nil {
			t.Error("purge error:", err)
		}

		queueJobId, _ := strconv.ParseUint(job.QueueJobId, 10, 0)
		_, err = client.StatsJob(queueJobId)
		if err == nil || err.Error() != gostalkc.NOT_FOUND {
			t.Error("Purge should have deleted job", job.QueueJobId)
		}
	}
}
