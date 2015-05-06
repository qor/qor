package integration

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

var DB gorm.DB

func Start(port int) {
	var err error
	// CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
	// CREATE DATABASE qor_integration_test;
	// GRANT ALL PRIVILEGES ON qor_integration_test.* TO 'qor'@'localhost';
	dbuser, dbpwd, dbname := "qor", "qor", "qor_integration_test"

	DB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))
	if err != nil {
		panic(err)
	}

	setupDb()

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

func setupDb() {
	tables := getTables()

	for _, table := range tables {
		if err := DB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}

		if err := DB.AutoMigrate(table).Error; err != nil {
			panic(err)
		}
	}
}
