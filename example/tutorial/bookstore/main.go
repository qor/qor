package main

import (
	"fmt"
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
	Admin.MountTo("/admin", mux)
	http.ListenAndServe(":9000", mux)
}
