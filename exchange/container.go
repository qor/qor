package exchange

import "github.com/qor/qor/resource"

type Container interface {
	Rows() (Rows, error)
}

type Rows interface {
	Columns() []string
	CurrentColumn() (*resource.MetaValues, error)
	Next() bool
}
