package main

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

func main() {
	Admin := admin.New(&qor.Config{DB: &DB})
	Admin.AddResource(
		&Author{},
		&admin.Config{Menu: []string{
			"Author Management"},
			Name: "Author",
		},
	)

	book := Admin.AddResource(
		&Book{},
		&admin.Config{
			Menu: []string{"Book Management"},
			Name: "Book",
		},
	)

	// step 5
	Admin.AddResource(
		&User{},
		&admin.Config{
			Menu: []string{"User Management"},
			Name: "User",
		},
	)

	book.Meta(&admin.Meta{Name: "Authors", Label: "Authors", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if authors := []Author{}; !context.GetDB().Find(&authors).RecordNotFound() {
				for _, author := range authors {
					results = append(results, []string{fmt.Sprintf("%v", author.ID), author.Name})
				}
			}
			return
		},
	})

	// Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})

	mux := http.NewServeMux()
	mux.HandleFunc("/books", func(res http.ResponseWriter, req *http.Request) {
		var books []Book
		DB.Find(&books)

		values := map[string]interface{}{
			"title": "List of Books",
			"books": books,
		}

		if tmpl, err := template.ParseFiles("templates/list.tmpl"); err == nil {
			tmpl.Execute(res, values)
		}
	})
	Admin.MountTo("/admin", mux)

	// Chapter 3: serve static files
	mux.Handle(
		"/system/",
		http.FileServer(http.Dir("public")),
	)

	// handle login and logout of users
	Admin.SetAuth(&Auth{})

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

	http.ListenAndServe(":9000", mux)
}
