package roles

import "testing"

func TestAllow(t *testing.T) {
	permission := Allow(Read, "api")

	if !permission.HasPermission(Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if permission.HasPermission(Update, "api") {
		t.Errorf("API should has no permission to Update")
	}
}

func TestCRUD(t *testing.T) {
	permission := Allow(CRUD, "admin")
	if !permission.HasPermission(Read, "admin") {
		t.Errorf("Admin should has permission to Read")
	}

	if !permission.HasPermission(Update, "admin") {
		t.Errorf("Admin should has permission to Update")
	}
}

func TestDeny(t *testing.T) {
	permission := Deny(Create, "api")

	if !permission.HasPermission(Read, "api") {
		t.Errorf("API should has permission to Read")
	}

	if !permission.HasPermission(Update, "api") {
		t.Errorf("API should has permission to Update")
	}

	if permission.HasPermission(Create, "api") {
		t.Errorf("API should has no permission to Update")
	}
}
