package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/worker"
)

var runWorker bool

func init() {
	flag.BoolVar(&runWorker, "run-worker", false, "run example beanstalkd worker")
	flag.Parse()
}

func main() {
	creditCard := admin.NewResource(CreditCard{})
	creditCard.Meta(&resource.Meta{Name: "issuer", Type: "select_one", Collection: []string{"VISA", "MasterCard", "UnionPay", "JCB", "American Express", "Diners Club"}})

	user := admin.NewResource(User{})
	user.IndexAttrs("fullname", "gender")
	user.Meta(&resource.Meta{Name: "CreditCard", Resource: creditCard})
	user.Meta(&resource.Meta{Name: "fullname", Alias: "name"})
	user.Meta(&resource.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.Meta(&resource.Meta{
		Name: "RoleId", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if roles := []Role{}; !context.GetDB().Find(&roles).RecordNotFound() {
				for _, role := range roles {
					results = append(results, []string{strconv.Itoa(role.Id), role.Name})
				}
			}
			return
		},
	})
	user.Meta(&resource.Meta{
		Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		},
	})

	config := qor.Config{DB: &db}
	web := admin.New(&config)
	web.UseResource(user)
	web.NewResource(Role{})
	web.NewResource(Language{})

	if runWorker {
		initWorkers(web)
	}

	fmt.Println("listening on :8080")
	mux := http.NewServeMux()
	web.MountTo("/admin", mux)
	http.ListenAndServe(":8080", mux)
}

func initWorkers(web *admin.Admin) {
	if err := worker.SetJobDB(&db); err != nil {
		panic(err)
	}

	worker.SetAdmin(web)

	bq := worker.NewBeanstalkdQueue("beanstalkd", "localhost:11300")
	var counter int
	w := worker.NewWorker(bq, "Log every 10 seconds", func(job *worker.Job) (err error) {
		counter++
		log, err := job.GetLogger()
		if err != nil {
			return
		}

		_, err = log.Write([]byte(strconv.Itoa(counter) + "\n"))

		time.Sleep(time.Minute * 5)

		return
	})

	// must be executed before http.ListenAndServer
	worker.Listen()

	job, err := w.NewJob(1, time.Now())
	if err != nil {
		panic(err)
	}
	_ = job
}
