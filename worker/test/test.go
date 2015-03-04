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
	db, err = gorm.Open("sqlite3", "tmp/worker.db")
	if err != nil {
		panic(err)
	}
	// db.LogMode(true)
}

func main() {
	config := qor.Config{DB: &db}
	web := admin.New(&config)
	// web.UseResource(user)

	// if err := db.DropTable(&worker.QorJob{}).Error; err != nil {
	// 	panic(err)
	// }
	if err := db.AutoMigrate(&worker.QorJob{}).Error; err != nil {
		panic(err)
	}
	worker.SetJobDB(&db)

	bq := worker.NewBeanstalkdQueue("beanstalkd", "localhost:11300")
	var counter int
	publishWorker := worker.New("Publish Jobs")

	web.AddResource(publishWorker, nil)

	publish := publishWorker.NewJob(bq, "publish products", "publish products so users could purchase new items", func(job *worker.QorJob) (err error) {
		log, err := job.GetLogger()
		if err != nil {
			return
		}

		_, err = log.Write([]byte(strconv.Itoa(counter) + "\n"))
		counter++
		time.Sleep(time.Minute * 5)
		return
	})

	// job.Meta(&admin.Meta{
	// 	Name: "File",
	// 	Type: "file",
	// 	Valuer: func(interface{}, *qor.Context) interface{} {
	// 		return nil
	// 	},
	// 	Setter: func(resource interface{}, metaValues *resource.MetaValues, context *qor.Context) {
	// 		return
	// 	},
	// })
	publish.Meta(&admin.Meta{
		Name: "Message",
		Type: "string",
	})
	publish.Meta(&admin.Meta{
		Name: "File",
		Type: "file",
	})

	publishWorker.NewJob(bq, "send mail magazines", "send mail magazines to subscribed users", nil)

	// extraInput := admin.NewResource(&Language{})
	// w.ExtraInput(extraInput)

	worker.Listen()

	// _ = job
	// if _, err := job.NewQorJob(1, time.Now()); err != nil {
	// 	panic(err)
	// }

	fmt.Println("listening on :8080")
	mux := http.NewServeMux()
	web.MountTo("/admin", mux)
	http.ListenAndServe(":8080", mux)
}
