package main

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/qor/qor/l10n"
	. "github.com/sclevine/agouti/matchers"
)

func TestL10n(t *testing.T) {
	SetupDb(true)
	defer StopDriverOnPanic()

	var productCN Product

	product := Product{Name: "Global product", Description: "Global product description", Code: "Global"}
	draftDB.Create(&product)

	Expect(page.Navigate(fmt.Sprintf("%v/product", baseUrl))).To(Succeed())

	cnProductName := "name for CN product"
	cnProductDesc := "description for CN product"
	cnProductL10nLink := fmt.Sprintf("a[href='/admin/product/%v?locale=zh-CN']", product.ID)
	Expect(page.Find(cnProductL10nLink).Click()).To(Succeed())

	page.Find("#QorResourceName").Fill(cnProductName)
	page.Find("#QorResourceDescription").Fill(cnProductDesc)
	// TODO: test product code should be disabled since it is `l10n:"sync"` attribute when this has been implemented

	page.FindByButton("Save").Click()

	if err := draftDB.Set("l10n:locale", "zh-CN").First(&productCN, product.ID).Error; err != nil {
		panic(err)
	}

	if productCN.Name != cnProductName {
		t.Error("cn product's description not set")
	}
	if productCN.Description != cnProductDesc {
		t.Error("cn product's description not set")
	}

	// Update global product, CN product's code should be changed too because it is `l10n:"sync"` attribute, but others should have no change
	Expect(page.Navigate(fmt.Sprintf("%v/product", baseUrl))).To(Succeed())
	Expect(page.Find(fmt.Sprintf("a[href='/admin/product/%v?locale=%v']", product.ID, l10n.Global)).Click()).To(Succeed())

	modifiedProductName := "modified name"
	modifiedProductDescription := "modified description"
	modifiedProductCode := "global 2"

	page.Find("#QorResourceName").Fill(modifiedProductName)
	page.Find("#QorResourceDescription").Fill(modifiedProductDescription)
	page.Find("#QorResourceCode").Fill(modifiedProductCode)
	page.FindByButton("Save").Click()

	draftDB.First(&product, product.ID)
	draftDB.Set("l10n:locale", "zh-CN").First(&productCN, product.ID)

	if product.Code != productCN.Code {
		t.Error("marked as sync attribute didn't change follow global change")
	}

	if productCN.Name == modifiedProductName || productCN.Description == modifiedProductDescription {
		t.Error("localized attribute has been changed")
	}
}

// Viewable locales []string{l10n.Global, "zh-CN", "JP", "EN", "DE"}
func TestL10nFilter(t *testing.T) {
	SetupDb(true)
	defer StopDriverOnPanic()

	product := Product{Name: "Global product", Description: "Global product description", Code: "Global"}
	draftDB.Create(&product)
	product.Name = "CN product"
	draftDB.Set("l10n:locale", "zh-CN").Create(&product)

	Expect(page.Navigate(fmt.Sprintf("%v/product", baseUrl))).To(Succeed())

	// Check l10n switcher
	Expect(page.Find(".lang-selector")).To(BeFound(), "locale selector not visible")

	viewableLocales := []string{l10n.Global, "zh-CN", "JP", "EN", "DE"}

	for _, locale := range viewableLocales {
		filterCSS := fmt.Sprintf(".lang-selector .dropdown-menu a[href='/admin/product?locale=%v']", locale)
		Expect(page.Find(filterCSS)).To(BeFound(), "locale selector not visible")
	}

	// Check global product
	Expect(page.Find(".lang-selector .dropdown-toggle").Text()).Should(BeEquivalentTo(l10n.Global), "Global locale isn't the default locale")
	page.Screenshot("/Users/raven1/Desktop/l10n.png")
	Expect(page.Find("td[title='Language Code']").Text()).Should(BeEquivalentTo(l10n.Global), "global product isn't visible")

	// Switch to zh-CN
	page.Find(".lang-selector .dropdown-toggle").Click()
	Expect(page.Find(fmt.Sprintf(".lang-selector .dropdown-menu a[href='/admin/product?locale=%v']", "zh-CN")).Click()).To(Succeed())

	Expect(page.Find(".lang-selector .dropdown-toggle").Text()).Should(BeEquivalentTo("zh-CN"), "zh-CN locale isn't visible in switcher")
	Expect(page.Find("td[title='Language Code']").Text()).Should(BeEquivalentTo("zh-CN"), "zh-CN product isn't visible")

	// Switch to JP
	page.Find(".lang-selector .dropdown-toggle").Click()
	Expect(page.Find(fmt.Sprintf(".lang-selector .dropdown-menu a[href='/admin/product?locale=%v']", "JP")).Click()).To(Succeed())

	Expect(page.Find("td[title='Language Code']").Text()).Should(BeEquivalentTo(l10n.Global), "unlocalized product doesn't show global language code")
}
