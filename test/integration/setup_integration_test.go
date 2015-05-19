package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

const (
	PORT = 4444
)

var (
	baseUrl = fmt.Sprintf("http://localhost:%v/admin", PORT)
	driver  *agouti.WebDriver
	page    *agouti.Page
)

func TestMain(m *testing.M) {
	var t *testing.T
	var err error

	driver = agouti.ChromeDriver()
	driver.Start()

	go Start(PORT)

	page, err = driver.NewPage()
	if err != nil {
		t.Error("Failed to open page.")
	}

	RegisterTestingT(t)
	test := m.Run()

	driver.Stop()
	os.Exit(test)
}

func StopDriverOnPanic() {
	var t *testing.T
	if r := recover(); r != nil {
		debug.PrintStack()
		fmt.Println("Recovered in f", r)
		driver.Stop()
		t.Fail()
	}
}
