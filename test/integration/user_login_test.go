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

	DB.First(&user)

	if user.Name != userName {
		t.Error("user not created")
	}
}
