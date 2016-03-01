package utils

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// TestDB initalize a db for testing
func TestDB() *gorm.DB {
	dbuser, dbpwd, dbname := "qor", "qor", "qor_test"

	if os.Getenv("TEST_ENV") == "CI" {
		dbuser, dbpwd = os.Getenv("DB_USER"), os.Getenv("DB_PWD")
	}

	var db gorm.DB
	var err error

	if os.Getenv("TEST_DB") == "postgres" {
		db, err = gorm.Open("postgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", dbuser, dbpwd, dbname))
	} else {
		// CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
		// CREATE DATABASE qor_test;
		// GRANT ALL ON qor_test.* TO 'qor'@'localhost';
		db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))
	}

	if err != nil {
		panic(err)
	}

	return &db
}
