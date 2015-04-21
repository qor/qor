package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

func main() {
	Admin := admin.New(&qor.Config{DB: Publish.DraftDB()})
	Admin.SetAuth(&Auth{})

	Admin.AddMenu(&admin.Menu{Name: "Dashboard", Link: "/admin"})

	creditcard := Admin.AddResource(&CreditCard{}, &admin.Config{Invisible: true})
	creditcard.Meta(&admin.Meta{Name: "issuer", Type: "select_one", Collection: []string{"VISA", "MasterCard", "UnionPay", "JCB", "American Express", "Diners Club"}})

	user := Admin.AddResource(&User{}, &admin.Config{Menu: []string{"Users Management"}})
	user.Meta(&admin.Meta{Name: "CreditCard", Resource: creditcard})
	user.Meta(&admin.Meta{Name: "fullname", Alias: "name"})

	user.Meta(&admin.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.Meta(&admin.Meta{Name: "RoleID", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if roles := []Role{}; !context.GetDB().Find(&roles).RecordNotFound() {
				for _, role := range roles {
					results = append(results, []string{fmt.Sprintf("%v", role.ID), role.Name})
				}
			}
			return
		},
	})

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

	user.Meta(&admin.Meta{Name: "description", Type: "rich_editor", Resource: Admin.NewResource(&admin.AssetManager{})})

	Admin.AddResource(&Language{}, &admin.Config{Name: "Locales", Menu: []string{"Products Management"}})
	product := Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Products Management"}})
	product.Meta(&admin.Meta{Name: "CollectionID", Label: "Collection", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if collections := []Collection{}; !context.GetDB().Find(&collections).RecordNotFound() {
				for _, collection := range collections {
					results = append(results, []string{fmt.Sprintf("%v", collection.ID), collection.Name})
				}
			}
			return
		},
	})
	Admin.AddResource(&Collection{}, &admin.Config{Menu: []string{"Products Management"}})
	Admin.AddResource(&Order{}, &admin.Config{Menu: []string{"Orders Management"}})

	Admin.AddResource(Publish)

	mux := http.NewServeMux()
	Admin.MountTo("/admin", mux)
	mux.Handle("/system/", http.FileServer(http.Dir("public")))
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var user User

		if request.Method == "POST" {
			request.ParseForm()
			if !DB.First(&user, "name = ?", request.Form.Get("username")).RecordNotFound() {
				cookie := http.Cookie{Name: "userid", Value: fmt.Sprintf("%v", user.ID), Expires: time.Now().AddDate(1, 0, 0)}
				http.SetCookie(writer, &cookie)
				writer.Write([]byte("<html><body>logged as `" + user.Name + "`, go <a href='/admin'>admin</a></body></html>"))
			} else {
				http.Redirect(writer, request, "/login?failed_to_login", 301)
			}
		} else if userid, err := request.Cookie("userid"); err == nil {
			if !DB.First(&user, "id = ?", userid.Value).RecordNotFound() {
				writer.Write([]byte("<html><body>already logged as `" + user.Name + "`, go <a href='/admin'>admin</a></body></html>"))
			} else {
				http.Redirect(writer, request, "/logout", http.StatusSeeOther)
			}
		} else {
			writer.Write([]byte(`<html><form action="/login" method="POST"><input name="username" value="" placeholder="username"><input type=submit value="Login"></form></html>`))
		}
	})

	mux.HandleFunc("/logout", func(writer http.ResponseWriter, request *http.Request) {
		cookie := http.Cookie{Name: "userid", MaxAge: -1}
		http.SetCookie(writer, &cookie)
		http.Redirect(writer, request, "/login?logged_out", http.StatusSeeOther)
	})

	fmt.Println("listening on :9000")
	http.ListenAndServe(":9000", mux)
}
