package sorting_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/sorting"
	"github.com/qor/qor/test/utils"
)

type User struct {
	gorm.Model
	Name string
	sorting.Sorting
}

var db *gorm.DB

func init() {
	db = utils.TestDB().Debug()
	sorting.RegisterCallbacks(db)
	db.DropTable(&User{})
	db.AutoMigrate(&User{})
}

func prepareUsers() {
	db.Delete(&User{})

	for i := 1; i <= 5; i++ {
		user := User{Name: fmt.Sprintf("user%v", i)}
		db.Save(&user)
	}
}

func getUser(name string) (user User) {
	db.First(&user, "name = ?", name)
	return user
}

func checkPosition(names ...string) bool {
	var users []User
	var positions []string

	db.Order("position").Find(&users)
	for _, user := range users {
		positions = append(positions, user.Name)
	}

	if reflect.DeepEqual(positions, names) {
		return true
	} else {
		fmt.Printf("Expect %v, got %v\n", names, positions)
		return false
	}
}

func TestMoveUpPosition(t *testing.T) {
	prepareUsers()
	sorting.MoveUp(db, getUser("user1"), 1)
	if !checkPosition("user2", "user1", "user3", "user4", "user5") {
		t.Errorf("user1 should be moved up")
	}

	sorting.MoveUp(db, getUser("user1"), 2)
	if !checkPosition("user2", "user3", "user4", "user1", "user5") {
		t.Errorf("user1 should be moved up")
	}

	sorting.MoveUp(db, getUser("user5"), 2)
	if !checkPosition("user2", "user3", "user4", "user1", "user5") {
		t.Errorf("user1 should be moved up")
	}
}
