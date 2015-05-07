package main

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
)

func TestForm(t *testing.T) {
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
	// File upload

	page.FindByButton("Save").Click()

	DB.Preload("Profile").First(&user)
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
}
