package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/qor/qor/worker"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
	"github.com/qor/qor/resource"
)

func main() {
	mux := http.NewServeMux()

	creditCard := admin.NewResource(CreditCard{})
	creditCard.Meta(&resource.Meta{Name: "issuer", Type: "select_one", Collection: []string{"VISA", "MasterCard", "UnionPay", "JCB", "American Express", "Diners Club"}})

	user := admin.NewResource(User{})
	user.IndexAttrs("fullname", "gender")
	user.Meta(&resource.Meta{Name: "CreditCard", Resource: creditCard})

	user.Meta(&resource.Meta{Name: "fullname", Alias: "name"})
	user.Meta(&resource.Meta{Name: "gender", Type: "select_one", Collection: []string{"M", "F", "U"}})
	user.Meta(&resource.Meta{Name: "RoleId", Label: "Role", Type: "select_one",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if roles := []Role{}; !context.GetDB().Find(&roles).RecordNotFound() {
				for _, role := range roles {
					results = append(results, []string{strconv.Itoa(role.Id), role.Name})
				}
			}
			return
		}})
	user.Meta(&resource.Meta{Name: "Languages", Type: "select_many",
		Collection: func(resource interface{}, context *qor.Context) (results [][]string) {
			if languages := []Language{}; !context.GetDB().Find(&languages).RecordNotFound() {
				for _, language := range languages {
					results = append(results, []string{strconv.Itoa(language.Id), language.Name})
				}
			}
			return
		}})

	config := qor.Config{DB: &db}
	web := admin.New(&config)
	web.UseResource(user)
	web.NewResource(Role{})
	web.NewResource(Language{})
	web.MountTo("/admin", mux)

	w := worker.New("Log every 10 seconds")

	worker.Listen()

	// go worker

	fmt.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
