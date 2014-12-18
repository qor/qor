package worker

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/manveru/gostalk/gostalkc"
)

type BeanstalkdQueue struct {
	addr string
	// db    *gorm.DB
	Debug io.Writer
}

// type BeanstalkdJob struct {
// 	Id              uint64
// 	QorJobId        uint64
// 	BeanstalkdJobId uint64
// }

func (bq *BeanstalkdQueue) newClient() (*gostalkc.Client, error) {
	return gostalkc.DialTimeout(bq.addr, time.Minute)
}

func (bq *BeanstalkdQueue) putJob(job *Job) (err error) {
	client, err := bq.newClient()
	if err != nil {
		return
	}

	delay := uint64(job.StartAt.Sub(time.Now()).Seconds())
	jobId := fmt.Sprintf("%d", job.Id)
	interval := strconv.FormatUint(job.Interval*60, 10)
	body := strings.Join([]string{jobId, interval}, ",")
	queueJobId, _, err := client.Put(0, delay, 60, []byte(body))
	if err != nil {
		return
	}

	// err = bq.db.Save(&BeanstalkdJob{QorJobId: job.Id, BeanstalkdJobId: queueJobId}).Error
	job.QueueJobId = fmt.Sprint(queueJobId)

	return
}

func NewBeanstalkdQueue(addr string) (bq *BeanstalkdQueue, err error) {
	bq = new(BeanstalkdQueue)
	bq.addr = addr
	// bq.db = db
	// err = bq.db.AutoMigrate(&BeanstalkdJob{}).Error

	return
}

func (bq *BeanstalkdQueue) Enqueue(job *Job) (err error) {
	return bq.putJob(job)
}

func (bq *BeanstalkdQueue) Dequeue() (jobId uint64, err error) {
	for {
		var workerConn *gostalkc.Client
		workerConn, err = bq.newClient()
		if err != nil {
			return
		}
		if err = workerConn.Conn.SetDeadline(time.Now().Add(time.Minute * 5)); err != nil {
			workerConn.Quit()

			if bq.Debug != nil {
				bq.Debug.Write([]byte("beanstalkd: failed to set deadline"))
			}

			continue
		}

		var body []byte
		var queueJobId uint64
		queueJobId, body, err = workerConn.ReserveWithTimeout(300)
		if err != nil {
			if err.Error() == gostalkc.TIMED_OUT {
				if bq.Debug != nil {
					bq.Debug.Write([]byte("beanstalkd: reserve timeout"))
				}
			} else if _, ok := err.(*net.OpError); ok {
				if bq.Debug != nil {
					bq.Debug.Write([]byte("beanstalkd: conn deadline"))
				}
			} else {
				return
			}

			continue
		}

		parts := bytes.Split(body, ",")
		jobId, err = strconv.ParseUint(string(parts[0]), 10, 0)
		if err != nil {
			return
		}

		if bq.Debug != nil {
			fmt.Fprintln(bq.Debug, "beanstalkd: receive job ", jobId)
		}

		if interval := string(parts[1]); interval == "0" {
			if err = workerConn.Delete(jobId); err != nil {
				return
			}
		} else {
			var i uint64
			i, err = strconv.ParseUint(interval, 10, 0)
			if err != nil {
				return
			}
			_, err = workerConn.Release(queueJobId, 0, i)
			if err != nil {
				return
			}
		}

		workerConn.Quit()

		return
	}

	return
}

func (bq *BeanstalkdQueue) Purge(job *Job) (err error) {
	workerConn, err := bq.newClient()
	if err != nil {
		return
	}
	jobId, err := strconv.ParseUint(job.QueueJobId, 10, 0)
	if err != nil {
		return
	}

	err = workerConn.Delete(jobId)

	return
}
