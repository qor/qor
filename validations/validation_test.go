package validations_test

import (
	"regexp"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/qor/qor/test/utils"
	"github.com/qor/qor/validations"
)

var db *gorm.DB

type User struct {
	gorm.Model
	Name       string
	CompanyID  int
	Company    Company
	CreditCard CreditCard
	Addresses  []Address
}

func (user *User) Validate(db *gorm.DB) {
	if user.Name == "" {
		validations.AddErrorForColumn(db, user, "Name", "invalid user name")
	}
}

type Company struct {
	gorm.Model
	Name string
}

func (company *Company) Validate(db *gorm.DB) {
	if company.Name == "" {
		validations.AddError(db, company, "invalid company name")
	}
}

type CreditCard struct {
	gorm.Model
	UserID int
	Number string
}

func (card *CreditCard) Validate(db *gorm.DB) {
	if regexp.MustCompile("^(\\d){13,16}$").MatchString(card.Number) {
		validations.AddErrorForColumn(db, card, "Number", "invalid card number")
	}
}

type Address struct {
	gorm.Model
	Address string
}

func (address *Address) Validate(db *gorm.DB) {
	if address.Address == "" {
		validations.AddErrorForColumn(db, address, "Address", "invalid address")
	}
}
func init() {
	db = utils.TestDB()
	validations.RegisterCallbacks(db)
	db.AutoMigrate(&User{}, &Company{}, &CreditCard{}, &Address{})
}

func TestSaveUesr(t *testing.T) {
	user := User{
		Company:    Company{},
		CreditCard: CreditCard{},
		Addresses:  []Address{{}},
	}

	if result := db.Debug().Save(&user); result.Error == nil {
		t.Errorf("Should get error when save blank user")
	}
}
