package exchange

import "github.com/qor/qor/resource"

type Container interface {
	Rows()
}

type Rows interface {
	Columns() []string
	CurrentColumn() *resource.MetaValues
	Next() bool
}
