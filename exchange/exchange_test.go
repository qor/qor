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
	Id        int64
	Name      string
	Age       int
	Addresses []Address
}

type Address struct {
	Id   int64
	Name string
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
	ex = New(testdb)
	userRes = ex.NewResource(User{})

	userRes.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	userRes.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})

	addRes := ex.NewResource(Address{})
	addRes.RegisterMeta(&resource.Meta{Name: "Name", Label: "Address1"})
	addRes.RegisterMeta(&resource.Meta{Name: "Name", Label: "Address2"})

	// ex.AddValidator(func(rel interface{}, mvs MetaValues, ctx *qor.Context) {
	// 	addMvs := mvs.Get("Addresses")
	// })

	// userRes.RegisterMeta(resource.Meta{Name: "xxx"})).Set("AutoCreate", true)
	// userRes.AddValidator(func(rel interface{}, mvs MetaValues, ctx *qor.Context) {})
}

func TestImport(t *testing.T) {
	r, err := os.Open("simple.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = userRes.Import(r, &qor.Context{DB: ex.DB})
	if err != nil {
		t.Fatal(err)
	}
	var users []User
	testdb.Find(&users)
	if len(users) != 3 {
		t.Fatalf("should get 3 records, but got %d", len(users))
	}
}

func TestMetaSet(t *testing.T) {
	res := ex.NewResource(User{})
	res.RegisterMeta(&resource.Meta{Name: "Name"}).Set("MultiDelimiter", ",").Set("HasSequentialColumns", true)
	meta := res.Metas["Name"].(*Meta)
	if meta.MultiDelimiter != "," {
		t.Errorf(`MultiDelimiter should be "," instead of "%s"`, meta.MultiDelimiter)
	}
	if !meta.HasSequentialColumns {
		t.Errorf(`MultiDelimiter should be "true" instead of "%s"`, meta.HasSequentialColumns)
	}
}
