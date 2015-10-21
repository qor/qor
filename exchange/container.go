package exchange

import "github.com/qor/qor/resource"

type Container interface {
	Rows() (Rows, error)
	WriteRow(*resource.MetaValues)
	WriteLog(string)
}

type Rows interface {
	Columns() []string
	CurrentColumn() (*resource.MetaValues, error)
	Next() bool
}
