package sorting_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/publish"
	"github.com/qor/qor/sorting"
	"github.com/qor/qor/test/utils"
)

type User struct {
	gorm.Model
	Name string
	sorting.Sorting
}

var db *gorm.DB
var pb *publish.Publish

func init() {
	db = utils.TestDB()
	sorting.RegisterCallbacks(db)

	pb = publish.New(db)
	pb.ProductionDB().DropTable(&User{}, &Product{})
	pb.DraftDB().DropTable(&Product{})
	db.AutoMigrate(&User{}, &Product{})
	pb.AutoMigrate(&Product{})
}

func prepareUsers() {
	db.Delete(&User{})

	for i := 1; i <= 5; i++ {
		user := User{Name: fmt.Sprintf("user%v", i)}
		db.Save(&user)
	}
}

func getUser(name string) *User {
	var user User
	db.First(&user, "name = ?", name)
	return &user
}

func checkPosition(names ...string) bool {
	var users []User
	var positions []string

	db.Find(&users)
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
	sorting.MoveUp(db, getUser("user5"), 2)
	if !checkPosition("user1", "user2", "user5", "user3", "user4") {
		t.Errorf("user5 should be moved up")
	}

	sorting.MoveUp(db, getUser("user5"), 1)
	if !checkPosition("user1", "user5", "user2", "user3", "user4") {
		t.Errorf("user5's postion won't be changed because it is already the last")
	}

	sorting.MoveUp(db, getUser("user1"), 1)
	if !checkPosition("user1", "user5", "user2", "user3", "user4") {
		t.Errorf("user1's position won't be changed because it is already on the top")
	}

	sorting.MoveUp(db, getUser("user5"), 1)
	if !checkPosition("user5", "user1", "user2", "user3", "user4") {
		t.Errorf("user5 should be moved up")
	}
}

func TestMoveDownPosition(t *testing.T) {
	prepareUsers()
	sorting.MoveDown(db, getUser("user1"), 1)
	if !checkPosition("user2", "user1", "user3", "user4", "user5") {
		t.Errorf("user1 should be moved down")
	}

	sorting.MoveDown(db, getUser("user1"), 2)
	if !checkPosition("user2", "user3", "user4", "user1", "user5") {
		t.Errorf("user1 should be moved down")
	}

	sorting.MoveDown(db, getUser("user5"), 2)
	if !checkPosition("user2", "user3", "user4", "user1", "user5") {
		t.Errorf("user5's postion won't be changed because it is already the last")
	}

	sorting.MoveDown(db, getUser("user1"), 1)
	if !checkPosition("user2", "user3", "user4", "user5", "user1") {
		t.Errorf("user1 should be moved down")
	}
}

func TestMoveToPosition(t *testing.T) {
	prepareUsers()
	user := getUser("user5")

	sorting.MoveTo(db, user, user.GetPosition()-3)
	if !checkPosition("user1", "user5", "user2", "user3", "user4") {
		t.Errorf("user5 should be moved to position 2")
	}

	sorting.MoveTo(db, user, user.GetPosition()-1)
	if !checkPosition("user5", "user1", "user2", "user3", "user4") {
		t.Errorf("user5 should be moved to position 1")
	}
}
