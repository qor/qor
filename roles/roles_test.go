package roles_test

import (
	"github.com/qor/qor"
	"github.com/qor/qor/roles"

	"testing"
)

func TestAllow(t *testing.T) {
	role := roles.New()
	role.Register("api", func(context *qor.Context) bool {
		return true
	})

	permission := role.Allow(roles.Read, "api")

	if !permission.HasPermission(roles.Read, &qor.Context{}) {
		t.Errorf("API should has permission to Read")
	}

	if permission.HasPermission(roles.Update, &qor.Context{}) {
		t.Errorf("API should has no permission to Update")
	}
}

func TestCURD(t *testing.T) {
	role := roles.New()
	role.Register("admin", func(context *qor.Context) bool {
		return true
	})

	permission := role.Allow(roles.CURD, "admin")
	if !permission.HasPermission(roles.Read, &qor.Context{}) {
		t.Errorf("Admin should has permission to Read")
	}

	if !permission.HasPermission(roles.Update, &qor.Context{}) {
		t.Errorf("Admin should has permission to Update")
	}
}

func TestDeny(t *testing.T) {
	role := roles.New()
	role.Register("api", func(context *qor.Context) bool {
		return true
	})

	permission := role.Deny(roles.Create, "api")

	if !permission.HasPermission(roles.Read, &qor.Context{}) {
		t.Errorf("API should has permission to Read")
	}

	if !permission.HasPermission(roles.Update, &qor.Context{}) {
		t.Errorf("API should has no permission to Update")
	}

	if permission.HasPermission(roles.Create, &qor.Context{}) {
		t.Errorf("API should has no permission to Update")
	}
}
