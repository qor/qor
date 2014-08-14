package exchange

import (
	"errors"
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
	cleanup()

	r, err := os.Open("simple.xlsx")
	if err != nil {
		t.Error(err)
	}
	fi, _, err := userRes.Import(r, &qor.Context{DB: ex.DB})
	if err != nil {
		t.Error(err)
	}

	if fi.TotalLines != 4 {
		t.Errorf("Total lines should be 4 instead of %d", fi.TotalLines)
	}

	select {
	case <-fi.Done:
	case err := <-fi.Error:
		t.Error(err)
	}

	var users []User
	testdb.Find(&users)
	if len(users) != 3 {
		t.Errorf("should get 3 records, but got %d", len(users))
	}
}

func cleanup() {
	testdb.DropTable(&User{})
	testdb.AutoMigrate(&User{})
	testdb.DropTable(&Address{})
	testdb.AutoMigrate(&Address{})
}

func TestImportError(t *testing.T) {
	cleanup()

	userRes := ex.NewResource(User{})
	userRes.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	userRes.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	ferr := errors.New("error")
	var i int
	userRes.AddValidator(func(rel interface{}, mvs resource.MetaValues, ctx *qor.Context) error {
		if i++; i == 2 {
			return ferr
		}
		return nil
	})

	r, err := os.Open("simple.xlsx")
	if err != nil {
		t.Error(err)
	}
	fi, iic, err := userRes.Import(r, &qor.Context{DB: ex.DB})
	if err != nil {
		t.Error(err)
	}

	hasError := true
	select {
	case <-fi.Done:
	case err := <-fi.Error:
		hasError = err != nil
	}

	if !hasError {
		t.Error("should return an error")
	}

	var j int
	var errs []error
	for ii := range iic {
		errs = append(errs, ii.Errors...)
		if j++; j == 3 {
			break
		}
	}

	if len(errs) != 1 && errs[0] != ferr {
		t.Error("Should receive errors properlly")
	}

	var users []User
	testdb.Find(&users)
	if len(users) != 0 {
		t.Errorf("should get 0 records, but got %d", len(users))
	}
}

func TestMetaSet(t *testing.T) {
	cleanup()

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
