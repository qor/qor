package exchange

import (
	"os"
	"testing"

	"github.com/qor/qor"
	"github.com/qor/qor/resource"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Id   int64
	Name string
	Age  int
}

var testdb = func() *gorm.DB {
	db, err := gorm.Open("sqlite3", "/tmp/qor_exchange_test.db")
	if err != nil {
		panic(err)
	}

	db.DropTable(&User{})
	db.AutoMigrate(&User{})

	return &db
}()

var (
	ex      *Exchange
	userRes *Resource
)

func init() {
	ex = &Exchange{DB: testdb}
	userRes = ex.NewResource(User{})
	userRes.RegisterMeta(&Meta{Meta: resource.Meta{Name: "Name", Label: "Name"}})
	userRes.RegisterMeta(&Meta{Meta: resource.Meta{Name: "Age", Label: "Age"}})
}

func TestImport(t *testing.T) {
	r, err := os.Open("simple.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	err = userRes.Import(r, &qor.Context{DB: ex.DB})
	if err != nil {
		t.Fatal(err)
	}
	var users []User
	testdb.Find(&users)
	if len(users) != 3 {
		t.Fatalf("should get 3 records, but got %d", len(users))
	}
}
