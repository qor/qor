package main

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func TestPublishProduct(t *testing.T) {
	SetupDb(true)
	defer StopDriverOnPanic()

	var product Product
	var publishedProduct Product

	// create product
	Expect(page.Navigate(fmt.Sprintf("%v/product", baseUrl))).To(Succeed())
	Expect(page.Find("#plus").Click()).To(Succeed())

	page.Find("#QorResourceName").Fill("product")
	page.Find("#QorResourceDescription").Fill("product")

	page.FindByButton("Save").Click()

	err := draftDB.Last(&product).Error
	notFoundErr := DB.Last(&publishedProduct).Error

	if err != nil {
		t.Error("new created resource not exists in draft table")
	}

	if notFoundErr == nil {
		t.Error("new created resource appears in production table")
	}

	// publish it
	Expect(page.Navigate(fmt.Sprintf("%v/publish", baseUrl))).To(Succeed())

	productCheckbox := fmt.Sprintf("#product__%v .selector input[type='checkbox']", product.ID)
	Expect(page.Find(productCheckbox).Check()).To(Succeed())

	Expect(page.FirstByButton("PUBLISH").Click()).To(Succeed())

	page.Session().AcceptAlert() // ConfirmPopup function doesn't work on CI. So use this function to confirm popup

	unpublishedCount, _ := page.Find(fmt.Sprintf("#product__%v", product.ID)).Count()
	if unpublishedCount != 0 {
		t.Error("smoke test, there should be no unpublished product")
	}

	draftDB.First(&product, product.ID)
	found := DB.Last(&publishedProduct).Error

	if found != nil {
		t.Error("published resource not saved in production table")
	}

	if publishedProduct.Name != product.Name || publishedProduct.Description != product.Description {
		t.Error("published resource has different content with draft resource")
	}
}

func TestDiscardProduct(t *testing.T) {
	SetupDb(true)
	defer StopDriverOnPanic()

	var product Product
	var publishedProduct Product

	// create product
	Expect(page.Navigate(fmt.Sprintf("%v/product", baseUrl))).To(Succeed())
	Expect(page.Find("#plus").Click()).To(Succeed())

	page.Find("#QorResourceName").Fill("product")
	page.Find("#QorResourceDescription").Fill("product")

	page.FindByButton("Save").Click()

	err := draftDB.Last(&product).Error
	notFoundErr := DB.Last(&publishedProduct).Error

	if err != nil {
		t.Error("new created resource not exists in draft table")
	}

	if notFoundErr == nil {
		t.Error("new created resource appears in production table")
	}

	// discard it
	Expect(page.Navigate(fmt.Sprintf("%v/publish", baseUrl))).To(Succeed())

	productCheckbox := fmt.Sprintf("#product__%v .selector input[type='checkbox']", product.ID)
	Expect(page.Find(productCheckbox).Check()).To(Succeed())

	Expect(page.FirstByButton("DISCARD").Click()).To(Succeed())
	Expect(page.ConfirmPopup()).To(Succeed())

	notExist := DB.Last(&publishedProduct).Error

	if notExist == nil {
		t.Error("discarded resource saved in production table")
	}
}
