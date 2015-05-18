package main

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
)

func TestCreateUser(t *testing.T) {
	defer StopDriverOnPanic()

	var user User
	userName := "test"

	if err := page.Navigate(baseUrl); err != nil {
		t.Error("Failed to navigate.")
	}

	Expect(page.Navigate(fmt.Sprintf("%v/user", baseUrl))).To(Succeed())
	Expect(page.Find("#plus").Click()).To(Succeed())
	Expect(page).To(HaveURL(fmt.Sprintf("%v/user/new", baseUrl)))

	page.Find("#QorResourceName").Fill(userName)
	page.FindByButton("Save").Click()

	DB.Last(&user)

	if user.Name != userName {
		t.Error("user not created")
	}
}

func TestUpdateUser(t *testing.T) {
	defer StopDriverOnPanic()

	user := &User{Name: "old name"}
	DB.Create(&user)
	newUserName := "new name"

	Expect(page.Navigate(fmt.Sprintf("%v/user", baseUrl))).To(Succeed())

	editLinkSelector := fmt.Sprintf("a[href='/admin/user/%v'].md-edit", user.ID)
	Expect(page.Find(editLinkSelector).Click()).To(Succeed())

	page.Find("#QorResourceName").Fill(newUserName)
	page.FindByButton("Save").Click()

	DB.First(&user, user.ID)

	if user.Name != newUserName {
		t.Error("user not updated")
	}
}

func TestDeleteUser(t *testing.T) {
	defer StopDriverOnPanic()

	user := &User{Name: "old name"}
	DB.Create(&user)

	Expect(page.Navigate(fmt.Sprintf("%v/user", baseUrl))).To(Succeed())

	deleteLinkSelector := fmt.Sprintf("form[action='/admin/user/%v'] button.md-delete", user.ID)
	Expect(page.Find(deleteLinkSelector).Click()).To(Succeed())
	alertErr := page.Session().AcceptAlert() // ConfirmPopup function doesn't work on CI. So use this function to confirm popup
	if alertErr != nil {
		t.Error("confirm box not accepted")
	}

	Eventually(page).Should(HaveURL(fmt.Sprintf("%v/user", baseUrl)))

	err := DB.First(&user, user.ID).Error

	if err == nil {
		t.Error("user not deleted")
	}
}
