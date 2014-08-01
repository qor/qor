package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
	"strconv"

	"net/http"
)

func main() {
	mux := http.NewServeMux()

	user := resource.New(&User{})
	user.Attrs().Index("name", "gender")
	user.Meta().Register(resource.Meta{Name: "File", Type: "file"})
	user.Meta().Register(resource.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.Meta().Register(resource.Meta{Name: "RoleId", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if roles := []Role{}; !context.DB.Find(&roles).RecordNotFound() {
				for _, role := range roles {
					results = append(results, []string{strconv.Itoa(role.Id), role.Name})
				}
			}
			return
		}})
	user.Meta().Register(resource.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.DB.Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		}})

	admin := admin.New(&db)
	admin.AddResource(user)
	admin.AddResource(resource.New(&Role{}))
	admin.AddResource(resource.New(&Language{}))
	admin.AddToMux("/admin", mux)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
