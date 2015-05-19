package utils

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func TestDB() gorm.DB {
	dbuser, dbpwd, dbname := "qor", "qor", "qor_test"

	if os.Getenv("TEST_ENV") == "CI" {
		dbuser, dbpwd = os.Getenv("DB_USER"), os.Getenv("DB_PWD")
	}
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))

	if err != nil {
		panic(err)
	}

	return db
}
