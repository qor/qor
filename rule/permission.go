package rule

const (
	READ   = 0
	WRITE  = 1
	RDWR   = 2
	CREATE = 3
	DELETE = 4
	ALL    = 5
)

type Permission struct {
}

func (p *Permission) HasPermission(mode int, i int) {
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
