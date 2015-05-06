package integration_test

import "testing"

func TestUserLoginPrompt(t *testing.T) {
	if err := page.Navigate(baseUrl); err != nil {
		t.Error("Failed to navigate.")
	}
}
