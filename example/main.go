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

	admin := admin.New(&db)

	user := admin.NewResource("users", User{})
	user.Attrs().Index("name", "gender")
	user.RegisterMeta(&resource.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.RegisterMeta(&resource.Meta{Name: "RoleId", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if roles := []Role{}; !context.DB.Find(&roles).RecordNotFound() {
				for _, role := range roles {
					results = append(results, []string{strconv.Itoa(role.Id), role.Name})
				}
			}
			return
		}})
	user.RegisterMeta(&resource.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.DB.Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		}})
	admin.NewResource("roles", Role{})
	admin.NewResource("languages", Language{})
	admin.AddToMux("/admin", mux)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
