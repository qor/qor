package main

import (
	"fmt"
	"net/http"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

func main() {
	devMode = true

	Start(3000)
}

func AdminConfig() (mux *http.ServeMux) {
	Admin := admin.New(&qor.Config{DB: &DB})

	user := Admin.AddResource(&User{})
	user.Meta(&admin.Meta{Name: "Gender", Type: "select_one", Collection: []string{"Male", "Female"}})
	user.Meta(&admin.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{fmt.Sprintf("%v", language.ID), language.Name})
				}
			}
			return
		},
	})
	user.Meta(&admin.Meta{Name: "Profile"})

	Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

	Admin.AddMenu(&admin.Menu{Name: "Google", Link: "http://www.google.com", Ancestors: []string{"Outside", "Search Engine"}})

	mux = http.NewServeMux()
	Admin.MountTo("/admin", mux)

	return
}

func Start(port int) {
	PrepareDB()
	mux := AdminConfig()
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
