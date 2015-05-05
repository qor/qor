package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
)

const (
	PORT = 9009
)

var _ = Describe("UserLogin", func() {
	var page *agouti.Page
	var baseUrl string

	BeforeEach(func() {
		go integraion.Start(PORT)

		baseUrl = fmt.Sprintf("http://localhost:%v/admin", PORT)
		var err error
		page, err = agoutiDriver.NewPage(agouti.Browser("chrome"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).To(Succeed())
	})

	It("should manage user authentication", func() {
		By("redirecting the user to the login form from the home page", func() {
			Expect(page.Navigate(baseUrl)).To(Succeed())
			Expect(page).To(HaveURL(baseUrl))
		})
	})
})
