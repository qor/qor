package rules_test

import (
	"github.com/qor/qor"
	"github.com/qor/qor/rules"

	"testing"
)

func TestAllow(t *testing.T) {
	rule := rules.New()
	rule.Register("api", func(context *qor.Context) bool {
		return true
	})

	permission := rule.Allow(rules.Read, "api")

	if !permission.HasPermission(rules.Read, &qor.Context{}) {
		t.Errorf("API should has permission to Read")
	}

	if permission.HasPermission(rules.Update, &qor.Context{}) {
		t.Errorf("API should has no permission to Update")
	}
}

func TestCURD(t *testing.T) {
	rule := rules.New()
	rule.Register("admin", func(context *qor.Context) bool {
		return true
	})

	permission := rule.Allow(rules.CURD, "admin")
	if !permission.HasPermission(rules.Read, &qor.Context{}) {
		t.Errorf("Admin should has permission to Read")
	}

	if !permission.HasPermission(rules.Update, &qor.Context{}) {
		t.Errorf("Admin should has permission to Update")
	}
}

func TestDeny(t *testing.T) {
	rule := rules.New()
	rule.Register("api", func(context *qor.Context) bool {
		return true
	})

	permission := rule.Deny(rules.Create, "api")

	if !permission.HasPermission(rules.Read, &qor.Context{}) {
		t.Errorf("API should has permission to Read")
	}

	if !permission.HasPermission(rules.Update, &qor.Context{}) {
		t.Errorf("API should has no permission to Update")
	}

	if permission.HasPermission(rules.Create, &qor.Context{}) {
		t.Errorf("API should has no permission to Update")
	}
}
