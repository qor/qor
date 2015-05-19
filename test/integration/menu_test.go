package main

import (
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/matchers"
)

// Menu structure: User | Product Management > Product | Outside > Search engine > Google
func TestMenu(t *testing.T) {
	defer StopDriverOnPanic()

	Expect(page.Navigate(baseUrl)).To(Succeed())

	// Have User menu at first level
	Expect(page.Find("div.navbar ul.nav li.item a[href='/admin/user']")).To(BeFound())

	// Have product menu at second level
	Expect(page.Find("div.navbar > ul.nav > li.item.dropdown > ul.dropdown-menu > li.item > a[href='/admin/product']")).To(BeFound())

	// Have google menu at third level
	Expect(page.Find("div.navbar > ul.nav > li.item.dropdown > ul.dropdown-menu > li.item > ul.dropdown-menu > li.item > a[href='http://www.google.com']")).To(BeFound())
}
