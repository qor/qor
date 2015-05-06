package integration_test

import (
	"fmt"
	"os"

	"testing"

	"github.com/qor/qor/test/integration"
	"github.com/sclevine/agouti"
)

const (
	PORT = 9009
)

var (
	baseUrl = fmt.Sprintf("http://localhost:%v/admin", PORT)
	driver  *agouti.WebDriver
	page    *agouti.Page
)

func TestMain(m *testing.M) {
	var t *testing.T
	var err error

	driver := agouti.Selenium()
	driver.Start()

	go integration.Start(PORT)

	page, err = driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		t.Error("Failed to open page.")
	}

	test := m.Run()

	driver.Stop()
	os.Exit(test)
}
