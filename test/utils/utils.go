package utils

import "testing"

// AssertNoErr makes current test case fatal if it receives a non-nil error
func AssertNoErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
