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

	creditCard := admin.NewResource(CreditCard{})
	creditCard.RegisterMeta(&resource.Meta{Name: "issuer", Type: "select_one", Collection: []string{"VISA", "MasterCard", "UnionPay", "JCB", "American Express", "Diners Club"}})

	user := admin.NewResource(User{})
	user.EditAttrs("fullname", "gender")
	user.RegisterMeta(&resource.Meta{Name: "CreditCard", Resource: creditCard})

	user.RegisterMeta(&resource.Meta{Name: "fullname", Alias: "name"})
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

	web := admin.New(&db)
	web.UseResource(user)
	web.NewResource(Role{})
	web.NewResource(Language{})
	web.AddToMux("/admin", mux)

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
