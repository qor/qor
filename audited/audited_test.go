package audited_test

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/audited"
	"github.com/qor/qor/test/utils"
)

type Product struct {
	gorm.Model
	Name string
	audited.AuditedModel
}

type User struct {
	gorm.Model
	Name string
}

var db *gorm.DB

func init() {
	db = utils.TestDB()
	db.AutoMigrate(&User{}, &Product{})
	audited.RegisterCallbacks(db)
}

func TestCreateUser(t *testing.T) {
	user := User{Name: "user1"}
	db.Save(&user)
	db := db.Set("qor:current_user", user)

	product := Product{Name: "product1"}
	db.Save(&product)
	if product.CreatedBy != fmt.Sprintf("%v", user.ID) {
		t.Errorf("created_by is not equal current user")
	}

	product.Name = "product_new"
	db.Save(&product)
	if product.UpdatedBy != fmt.Sprintf("%v", user.ID) {
		t.Errorf("updated_by is not equal current user")
	}
}
