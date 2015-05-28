package main

import (
	"fmt"
	"html/template"
	"net/http"

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
	http.ListenAndServe(":9000", mux)
}
