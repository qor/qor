package exchange

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type User struct {
	Id           int64
	Name         string
	Age          int
	CellPhone    Phone
	Addresses    []Address
	OldAddresses []Address
}

type Address struct {
	Id      int64
	UserId  int64
	Name    string
	Country string

	Phone Phone
}

type Phone struct {
	Id        int64
	UserId    int64
	AddressId int64
	Num       string
	CreatedAt time.Time
}

var (
	testdb = func() *gorm.DB {
		db, err := gorm.Open("sqlite3", "/tmp/qor_exchange_test.db")
		if err != nil {
			panic(err)
		}

		return &db
	}()
)

func cleanup() {
	testdb.DropTable(&User{})
	testdb.AutoMigrate(&User{})
	testdb.DropTable(&Address{})
	testdb.AutoMigrate(&Address{})
}

func TestImportSimple(t *testing.T) {
	cleanup()

	useres := NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	ex := New(useres)

	f, err := NewXLSXFile("simple.xlsx")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = ex.Import(f, &buf, &qor.Context{Config: &qor.Config{DB: testdb}})
	if err != nil {
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

	useres := NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	addres := NewResource(Address{})
	addres.HasSequentialColumns = true
	useres.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: addres})
	addres.RegisterMeta(&resource.Meta{Name: "Country", Label: "Country"})
	ex := New(useres)

	f, err := NewXLSXFile("nested_resource.xlsx")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = ex.Import(f, &buf, &qor.Context{Config: &qor.Config{DB: testdb}})
	if err != nil {
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

type FullMarathon struct {
	RunningLevel float64
	Min1500      int
	Sec1500      int
}

func TestImportNormalizeHeader(t *testing.T) {
	testdb.DropTable(FullMarathon{})
	testdb.AutoMigrate(FullMarathon{})

	marathon := NewResource(FullMarathon{})
	marathon.RegisterMeta(&resource.Meta{Name: "RunningLevel", Label: "Running Level"})
	marathon.RegisterMeta(&resource.Meta{Name: "Min1500", Label: "1500M Min"})
	marathon.RegisterMeta(&resource.Meta{Name: "Sec1500", Label: "1500M Sec"})
	ex := New(marathon)
	ex.JobThrottle = 10
	ex.DataStartAt = 3
	ex.NormalizeHeaders = func(f File) (headers []string) {
		if f.TotalLines() < 3 {
			return
		}

		topHeaders, subHeaders := f.Line(1), f.Line(2)
		if len(topHeaders) != len(subHeaders) {
			return
		}
		for i, subHeader := range subHeaders {
			var prefix string
			topSec := topHeaders[:i+1]
			lenSec := len(topSec)
			for j := lenSec - 1; j >= 0; j-- {
				if topHeader := topSec[j]; topHeader != "" {
					prefix = topHeader
					break
				}
			}

			headers = append(headers, prefix+" "+subHeader)
		}

		return
	}
	f, err := NewXLSXFile("headers.xlsx")
	if err != nil {
		t.Error(err)
	}

	var buf bytes.Buffer
	err = ex.Import(f, &buf, &qor.Context{Config: &qor.Config{DB: testdb}})
	if err != nil {
		t.Error(err)
	}

	var marathones []FullMarathon
	testdb.Find(&marathones)
	if len(marathones) != 12 {
		t.Errorf("should get 12 records, but got %d", len(marathones))
	}
	if marathones[1].RunningLevel != 28.1 {
		t.Errorf("should get 28.1, but got %f", marathones[1].RunningLevel)
	}
	if marathones[1].Min1500 != 8 {
		t.Errorf("should get 8, but got %f", marathones[1].Min1500)
	}
	if marathones[1].Sec1500 != 26 {
		t.Errorf("should get 26, but got %f", marathones[1].Sec1500)
	}
}

func TestImportError(t *testing.T) {
	cleanup()

	useres := NewResource(User{})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	ex := New(useres)
	ferr := errors.New("an validator error in the second line")
	var i int
	useres.AddValidator(func(rel interface{}, mvs *resource.MetaValues, ctx *qor.Context) error {
		if i++; i == 2 {
			return ferr
		}
		return nil
	})

	f, err := NewXLSXFile("simple.xlsx")
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	err = ex.Import(f, &buf, &qor.Context{Config: &qor.Config{DB: testdb}})
	if err == nil {
		t.Error("Should encouter error")
	}

	logs := "2/4: 1 Van 24  \n3/4: an validator error in the second line; \n4/4: 3 Kate 25  \n"
	if buf.String() != logs {
		t.Errorf(`Expect log %q but got %q`, logs, buf.String())
	}

	var users []User
	testdb.Find(&users)
	if len(users) != 0 {
		t.Errorf("should get 0 records, but got %d", len(users))
	}
}

func TestExport(t *testing.T) {
	records := []interface{}{
		User{
			Name: "Van",
			Age:  24,
			Addresses: []Address{
				{Country: "China", Phone: Phone{Num: "xxx-xxx-xxx-0"}},
				{Country: "Japan", Phone: Phone{Num: "zzz-zzz-zzz-0"}},
				{Country: "New Zealand", Phone: Phone{Num: "kkk-kkk-kkk-0"}},
			},
			OldAddresses: []Address{
				{Country: "Africa"},
			},
			CellPhone: Phone{Num: "yyy-yyy-yyy-0"},
		},
		User{
			Name: "Jane",
			Age:  26,
			Addresses: []Address{
				{Country: "USA", Phone: Phone{Num: "xxx-xxx-xxx-1"}},
			},
			OldAddresses: []Address{
				{Country: "Africa"},
				{Country: "Brazil"},
			},
			CellPhone: Phone{Num: "yyy-yyy-yyy-1"},
		},
	}

	phone := NewResource(Phone{})
	cellphone := NewResource(Phone{})
	addres := NewResource(Address{})
	oldaddres := NewResource(Address{})
	useres := NewResource(User{})
	phone.HasSequentialColumns = true
	addres.HasSequentialColumns = true
	oldaddres.MultiDelimiter = ","

	phone.RegisterMeta(&resource.Meta{Name: "Num", Label: "Phone"})
	oldaddres.RegisterMeta(&resource.Meta{Name: "Country", Label: "Old Countries"})
	cellphone.RegisterMeta(&resource.Meta{Name: "Num", Label: "CellPhone"})

	addres.RegisterMeta(&resource.Meta{Name: "Phone", Resource: phone})
	addres.RegisterMeta(&resource.Meta{Name: "Country", Label: "Country"})

	useres.RegisterMeta(&resource.Meta{Name: "OldAddresses", Resource: oldaddres})
	useres.RegisterMeta(&resource.Meta{Name: "Addresses", Resource: addres})
	useres.RegisterMeta(&resource.Meta{Name: "CellPhone", Resource: cellphone})
	useres.RegisterMeta(&resource.Meta{Name: "Age", Label: "Age"})
	useres.RegisterMeta(&resource.Meta{Name: "Name", Label: "Name"})

	ex := New(useres)
	var buf bytes.Buffer
	ex.Export(records, &buf, &qor.Context{Config: &qor.Config{DB: testdb}})
	expect := `Old Countries,Phone 01,Phone 02,Phone 03,Country 01,Country 02,Country 03,CellPhone,Age,Name
Africa,xxx-xxx-xxx-0,zzz-zzz-zzz-0,kkk-kkk-kkk-0,China,Japan,New Zealand,yyy-yyy-yyy-0,24,Van
"Africa,Brazil",xxx-xxx-xxx-1,,,USA,,,yyy-yyy-yyy-1,26,Jane
`

	if buf.String() != expect {
		t.Errorf("expect: %q\ngot: %q", expect, buf.String())
	}
}
