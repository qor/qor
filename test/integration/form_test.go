package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
)

func TestForm(t *testing.T) {
	SetupDb(true)
	defer StopDriverOnPanic()

	var user User
	var languages []Language
	userName := "user name"
	address := "an address"

	langEN := &Language{Name: "en"}
	langCN := &Language{Name: "cn"}
	DB.Create(&langEN)
	DB.Create(&langCN)

	Expect(page.Navigate(fmt.Sprintf("%v/user", baseUrl))).To(Succeed())
	Expect(page.Find("#plus").Click()).To(Succeed())
	Expect(page).To(HaveURL(fmt.Sprintf("%v/user/new", baseUrl)))

	// Text input
	page.Find("#QorResourceName").Fill(userName)

	// Select one
	page.Find("#QorResourceGender_chosen").Click()
	page.Find("#QorResourceGender_chosen .chosen-drop ul.chosen-results li[data-option-array-index='1']").Click()

	// Select many
	page.Find("#QorResourceLanguages_chosen .search-field input").Click()
	page.Find("#QorResourceLanguages_chosen .chosen-drop ul.chosen-results li[data-option-array-index='0']").Click()

	page.Find("#QorResourceLanguages_chosen").Click()
	Expect(page.Find("#QorResourceLanguages_chosen .chosen-drop ul.chosen-results li[data-option-array-index='1']").Click()).To(Succeed())

	// Nested resource
	page.Find("#QorResourceProfileAddress").Fill(address)

	// Rich text
	Expect(page.Find(".redactor-box")).To(BeFound())

	// File upload
	Expect(page.Find("input[name='QorResource.Avatar']").UploadFile("fixtures/ThePlant.png")).To(Succeed())

	page.FindByButton("Save").Click()

	DB.Preload("Profile").Last(&user)
	DB.Model(&user).Related(&languages, "Languages")

	if user.Name != userName {
		t.Error("text input for user name not work")
	}

	if user.Gender != "Male" {
		t.Error("select_one for gender not work")
	}

	if len(languages) != 2 {
		t.Error("select_many for languages not work")
	}

	if user.Profile.Address != address {
		t.Error("nested resource for profile not work")
	}

	avatarFile := fmt.Sprintf("public%v", user.Avatar.Url)
	if _, err := os.Stat(avatarFile); os.IsNotExist(err) {
		t.Error("file uploader for avatar not work")
	} else {
		os.Remove(avatarFile)
		// Remove uploaded .original file
		// File path looks like public/system/users/1/Avatar/ThePlant20150508172715879986152.original.png
		filePaths := strings.Split(avatarFile, ".")
		os.Remove(fmt.Sprintf("%v.%v.original.%v", filePaths[0], filePaths[1], filePaths[2]))
	}
}
