package main

import (
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"

	"net/http"
)

func main() {
	mux := http.NewServeMux()

	admin := admin.New(&db)

	user := admin.NewResource("user", User{})
	user.IndexAttrs("name", "gender")
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
	admin.NewResource("role", Role{})
	admin.NewResource("language", Language{})
	admin.AddToMux("/admin", mux)

	// exchanger := exchange.New(&db)
	// userexchanger := exchange.NewResource(&CreditCard{})
	// ccexchanger.RegisterMeta(&exchange.Meta{Name: "Number", Label: "CC Number"})
	// userexchanger.RegisterMeta(exchange.Meta{Name: "CreditCard", Resource: ccexchanger})
	// exchanger.UseResource(userexchanger)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
