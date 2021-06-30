package utils

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

// TestDB initialize a db for testing
func TestDB() *gorm.DB {
	var db *gorm.DB
	var err error
	var dbuser, dbpwd, dbname, dbhost = "qor", "qor", "qor_test", "localhost"

	if os.Getenv("DB_USER") != "" {
		dbuser = os.Getenv("DB_USER")
	}

	if os.Getenv("DB_PWD") != "" {
		dbpwd = os.Getenv("DB_PWD")
	}

	if os.Getenv("DB_NAME") != "" {
		dbname = os.Getenv("DB_NAME")
	}

	if os.Getenv("DB_HOST") != "" {
		dbhost = os.Getenv("DB_HOST")
	}

	if os.Getenv("TEST_DB") == "mysql" {
		// CREATE USER 'qor'@'localhost' IDENTIFIED BY 'qor';
		// CREATE DATABASE qor_test;
		// GRANT ALL ON qor_test.* TO 'qor'@'localhost';
		db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", dbuser, dbpwd, dbname))
	} else {
		db, err = gorm.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbuser, dbpwd, dbhost, dbname))
	}

	if err != nil {
		panic(err)
	}

	if os.Getenv("DEBUG") != "" {
		db.LogMode(true)
	}

	return db
}

var db *gorm.DB

func GetTestDB() *gorm.DB {
	if db != nil {
		return db
	}

	db = TestDB()

	return db
}

// PrepareDBAndTables prepare given tables cleanly and return a test database instance
func PrepareDBAndTables(tables ...interface{}) *gorm.DB {
	db := GetTestDB()

	ResetDBTables(db, tables...)

	return db
}

// ResetDBTables reset given tables.
func ResetDBTables(db *gorm.DB, tables ...interface{}) {
	Truncate(db, tables...)
	AutoMigrate(db, tables...)
}

// Truncate receives table arguments and truncate their content in database.
func Truncate(db *gorm.DB, givenTables ...interface{}) {
	// We need to iterate throught the list in reverse order of
	// creation, since later tables may have constraints or
	// dependencies on earlier tables.
	len := len(givenTables)
	for i := range givenTables {
		table := givenTables[len-i-1]
		db.DropTableIfExists(table)
	}
}

// AutoMigrate receives table arguments and create or update their
// table structure in database.
func AutoMigrate(db *gorm.DB, givenTables ...interface{}) {
	for _, table := range givenTables {
		db.AutoMigrate(table)
		if migratable, ok := table.(Migratable); ok {
			exec(func() error { return migratable.AfterMigrate(db) })
		}
	}
}

// Migratable defines interface for implementing post-migration
// actions such as adding constraints that arent's supported by Gorm's
// struct tags. This function must be idempotent, since it will most
// likely be executed multiple times.
type Migratable interface {
	AfterMigrate(db *gorm.DB) error
}

func exec(c func() error) {
	if err := c(); err != nil {
		panic(err)
	}
}
