package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/qor/qor/l10n"
	"github.com/qor/qor/media_library"
)

type User struct {
	gorm.Model
	Name      string
	Gender    string
	Languages []Language `gorm:"many2many:user_languages;"`
	Note      string
	Avatar    media_library.FileSystem

	Profile Profile
}

type Profile struct {
	gorm.Model
	UserId  uint64
	Address string
}

type Language struct {
	gorm.Model
	Name string
}

type Product struct {
	gorm.Model
	Name        string
	Description string `l10n:"sync"`
	l10n.Locale
}

var (
	DB      gorm.DB
	devMode bool
	dbname  string
	dbuser  string
	dbpwd   string
)

func PrepareDB() {
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

	SetupDb(!devMode) // Don't drop table in dev mode

	Login()

	l10n.RegisterCallbacks(&DB)
}

func getTables() []interface{} {
	return []interface{}{
		&User{},
		&Product{},
		&Profile{},
		&Language{},
	}
}

func SetupDb(dropBeforeCreate bool) {
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

func (User) ViewableLocales() []string {
	return []string{l10n.Global, "zh-CN", "JP", "EN", "DE"}
}

func (user User) EditableLocales() []string {
	return []string{l10n.Global, "zh-CN", "EN"}
}

func (u User) DisplayName() string {
	return u.Name
}

// Set current user via db directly. see auth.go for detail. For test only
func Login() {
	currentUser := User{Name: "currentUser"}

	if DB.Where("name = ?", "currentUser").First(&currentUser).RecordNotFound() {
		if err := DB.Create(&currentUser).Error; err != nil {
			panic(err)
		}
	}
}
