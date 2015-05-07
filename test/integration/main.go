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

	Admin.AddResource(&User{})
	Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

	Admin.AddMenu(&admin.Menu{Name: "Google", Link: "http://www.google.com", Ancestors: []string{"Outside", "Search Engine"}})

	mux = http.NewServeMux()
	Admin.MountTo("/admin", mux)

	return
}

func Start(port int) {
	mux := AdminConfig()
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
