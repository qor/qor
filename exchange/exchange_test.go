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
	Id      int64
	UserId  int64
	Name    string
	Country string
}

var (
	testdb = func() *gorm.DB {
		db, err := gorm.Open("sqlite3", "/tmp/qor_exchange_test.db")
		if err != nil {
			panic(err)
		}

		return &db
	}()
	ex = New()
)

func TestImportSimple(t *testing.T) {
	cleanup()

	useres := ex.NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})

	r, err := os.Open("simple.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	fi, _, err := useres.Import(r, &qor.Context{DB: ex.DB})
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

func TestImportNested(t *testing.T) {
	cleanup()

	useres := ex.NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	addres := ex.NewResource(Address{})
	addres.HasSequentialColumns = true
	useres.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: addres})
	addres.RegisterMeta(&resource.Meta{Name: "Country", Label: "Address"})

	r, err := os.Open("nested_resource.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	fi, _, err := useres.Import(r, &qor.Context{DB: ex.DB})
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
		t.Errorf("should get 3 users, but got %d", len(users))
	}
	var addresses []Address
	testdb.Find(&addresses)
	if len(addresses) != 6 {
		t.Errorf("should get 6 addresses, but got %d", len(addresses))
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

	useres := ex.NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	ferr := errors.New("error")
	var i int
	useres.AddValidator(func(rel interface{}, mvs *resource.MetaValues, ctx *qor.Context) error {
		if i++; i == 2 {
			return ferr
		}
		return nil
	})

	r, err := os.Open("simple.xlsx")
	if err != nil {
		t.Error(err)
	}
	fi, iic, err := useres.Import(r, &qor.Context{DB: ex.DB})
	if err != nil {
		t.Error(err)
	}

	var hasError bool
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

func TestGetMetaValues(t *testing.T) {
	useres := ex.NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	addres := ex.NewResource(Address{})
	addres.HasSequentialColumns = true
	useres.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: addres})
	addres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Address"})

	mvs := useres.GetMetaValues(map[string]string{
		"Name":       "Van",
		"Address 01": "China",
		"Address 02": "USA",
	}, 0)

	if len(mvs.Values) != 4 {
		t.Errorf("expecting to retrieve 4 MetaValues instead of %d", len(mvs.Values))
	}

	var hasChina, hasUSA bool
	for _, v := range mvs.Values {
		if v.MetaValues == nil {
			continue
		}
		switch v.MetaValues.Values[0].Value.(string) {
		case "China":
			hasChina = true
		case "USA":
			hasUSA = true
		}
	}

	if !hasChina {
		t.Error("Should contains China in mvs.Values")
	}
	if !hasUSA {
		t.Error("Should contains USA in mvs.Values")
	}
}
