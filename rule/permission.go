package rule

type PermissionMode uint32

const (
	Read PermissionMode = 1 << (32 - 1 - iota)
	Update
	Create
	Delete
	All
)

type Permission struct {
	allowRoles map[int]string
	denyRoles  map[int]string
}

func (p *Permission) HasPermission(mode int, role string) bool {
	return false
}

func Allow(mode int, roles ...string) *Permission {
	permission := &Permission{}
	permission.Allow(mode, roles...)
	return permission
}

func (p *Permission) Allow(mode int, roles ...string) *Permission {
	return p
}

func Deny(mode int, roles ...string) *Permission {
	permission := &Permission{}
	permission.Deny(mode, roles...)
	return permission
}

func (p *Permission) Deny(mode int, roles ...string) *Permission {
	return p
}
