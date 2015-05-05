package integration

import (
	"fmt"
	"net/http"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
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

func Start(port int) {
	DB, _ := gorm.Open("sqlite3", "demo.db")
	DB.AutoMigrate(&User{}, &Product{})

	Admin := admin.New(&qor.Config{DB: &DB})
	Admin.AddResource(&User{}, &admin.Config{Menu: []string{"User Management"}})
	Admin.AddResource(&Product{}, &admin.Config{Menu: []string{"Product Management"}})

	mux := http.NewServeMux()
	Admin.MountTo("/admin", mux)
	http.ListenAndServe(fmt.Sprintf(":%v", port), mux)
}
