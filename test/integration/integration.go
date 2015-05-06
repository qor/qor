package main

import (
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor"
	"github.com/qor/qor/admin"
)

type User struct {
	gorm.Model
	Name string
}

type Product struct {
	gorm.Model
	Name        string
	Description string
}

var (
	DB      gorm.DB
	devMode bool
	dbname  string
	dbuser  string
	dbpwd   string
)

func main() {
	devMode = true

	Start(3000)
}

func Start(port int) {
	var err error

	// Be able to start a server for develop test
	dbuser, dbpwd = "qor", "qor"
	if devMode {
		dbname = "qor_integration"
	} else {
		dbname = "qor_integration_test"
	}

	DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))
	if err != nil {
		panic(err)
	}

	setupDb(!devMode) // Don't drop table in dev mode

	Admin := admin.New(&qor.Config{DB: &DB})

	Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

	mux := http.NewServeMux()
	Admin.MountTo("/admin", mux)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}

func getTables() []interface{} {
	return []interface{}{
		&User{},
		&Product{},
	}
}

func setupDb(dropBeforeCreate bool) {
	tables := getTables()

	for _, table := range tables {
		if dropBeforeCreate {
			if err := DB.DropTableIfExists(table).Error; err != nil {
				panic(err)
			}
		}

		if err := DB.AutoMigrate(table).Error; err != nil {
			panic(err)
		}
	}
}
