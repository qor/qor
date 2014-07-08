package rules_test

import (
	"fmt"
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"testing"
)

func TestAddRule(t *testing.T) {
	rule := rules.New()
	rule.Register("admin", func(context *qor.Context) bool {
		return true
	})

	permission := rule.Allow(rules.Read, "admin", "api")
	fmt.Println(permission.HasPermission(rules.Update, &qor.Context{}))
}
