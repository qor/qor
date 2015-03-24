package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

var runWorker bool

func init() {
	flag.BoolVar(&runWorker, "run-worker", false, "run example beanstalkd worker")
	flag.Parse()
}

func main() {
	config := qor.Config{DB: Publish.DraftDB()}

	Admin := admin.New(&config)

	Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

	creditCard := Admin.AddResource(&CreditCard{}, &admin.Config{Menu: []string{"User Management"}})

	creditCard.Meta(&admin.Meta{Name: "issuer", Type: "select_one", Collection: []string{"VISA", "MasterCard", "UnionPay", "JCB", "American Express", "Diners Club"}})

	user := Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
	user.Meta(&admin.Meta{Name: "CreditCard", Resource: creditCard})
	user.Meta(&admin.Meta{Name: "fullname", Alias: "name"})

	user.Meta(&admin.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.Meta(&admin.Meta{Name: "RoleID", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if roles := []Role{}; !context.GetDB().Find(&roles).RecordNotFound() {
				for _, role := range roles {
					results = append(results, []string{strconv.Itoa(role.ID), role.Name})
				}
			}
			return
		},
	})

	user.Meta(&admin.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.ID), language.Name})
				}
			}
			return
		},
	})

	Admin.AddResource(&Language{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

	assetManager := Admin.AddResource(&admin.AssetManager{}, nil)
	user.Meta(&admin.Meta{Name: "description", Type: "rich_editor", Resource: assetManager})

	Admin.AddResource(Publish, nil)

	mux := http.NewServeMux()
	Admin.MountTo("/admin", mux)
	mux.Handle("/system/", http.FileServer(http.Dir("public")))
	mux.HandleFunc("/login", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "POST" {
			request.ParseForm()
			var user User
			if !DB.First(&user, "name = ?", request.Form.Get("username")).RecordNotFound() {
				loggedUserId = user.ID
				writer.Write([]byte("logged as " + user.Name))
			} else {
				http.Redirect(writer, request, "/login?failed_to_login", 301)
			}
		} else if loggedUserId != 0 {
			var user User
			DB.First(&user, "id = ?", loggedUserId)
			writer.Write([]byte("already logged as " + user.Name))
		} else {
			writer.Write([]byte(`<html><form action="/login" method="POST"><input name="username" value="" placeholder="username"><input type=submit value="Login"></form></html>`))
		}
	})

	mux.HandleFunc("/logout", func(writer http.ResponseWriter, request *http.Request) {
		loggedUserId = 0
		http.Redirect(writer, request, "/login?logged_out", http.StatusSeeOther)
	})

	fmt.Println("listening on :9000")
	http.ListenAndServe(":9000", mux)
}
