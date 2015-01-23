package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/worker"
)

var db gorm.DB

func init() {
	var err error
	db, err = gorm.Open("sqlite3", "worker.db")
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
}

func main() {
	config := qor.Config{DB: &db}
	web := admin.New(&config)
	// web.UseResource(user)

	if err := worker.SetJobDB(&db); err != nil {
		panic(err)
	}

	bq := worker.NewBeanstalkdQueue("beanstalkd", "localhost:11300")
	var counter int
	publishWorker := worker.New("publish_jobs")

	web.NewResource(publishWorker, publishWorker.Name)

	publishWorker.NewJob(bq, "publish products", func(job *worker.QorJob) (err error) {
		log, err := job.GetLogger()
		if err != nil {
			return
		}

		_, err = log.Write([]byte(strconv.Itoa(counter) + "\n"))
		counter++
		time.Sleep(time.Minute * 5)
		return
	})

	// extraInput := admin.NewResource(&Language{})
	// w.ExtraInput(extraInput)

	worker.Listen()

	// if _, err := w.NewJob(1, time.Now()); err != nil {
	// 	panic(err)
	// }

	fmt.Println("listening on :8080")
	mux := http.NewServeMux()
	web.MountTo("/admin", mux)
	http.ListenAndServe(":8080", mux)
}
