package roles_test

import (
	"github.com/qor/qor/roles"

	"testing"
)

func TestAllow(t *testing.T) {
	permission := roles.Allow(roles.Read, "api")

	if !permission.HasPermission(roles.Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if permission.HasPermission(roles.Update, "api") {
		t.Errorf("API should has no permission to Update")
	}
}

func TestCRUD(t *testing.T) {
	permission := roles.Allow(roles.CRUD, "admin")
	if !permission.HasPermission(roles.Read, "admin") {
		t.Errorf("Admin should has permission to Read")
	}

	if !permission.HasPermission(roles.Update, "admin") {
		t.Errorf("Admin should has permission to Update")
	}
}

func TestDeny(t *testing.T) {
	permission := roles.Deny(roles.Create, "api")

	if !permission.HasPermission(roles.Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if !permission.HasPermission(roles.Update, "api") {
		t.Errorf("API should has no permission to Update")
	}

	if permission.HasPermission(roles.Create, "api") {
		t.Errorf("API should has no permission to Update")
	}
}
