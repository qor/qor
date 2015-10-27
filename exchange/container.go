package exchange

import (
	"github.com/qor/qor"
	"github.com/qor/qor/resource"
)

type Container interface {
	NewReader(*Resource, *qor.Context) (Rows, error)
	NewWriter(*Resource, *qor.Context) (Writer, error)
	WriteLog(string)
}

type Rows interface {
	Header() []string
	ReadRow() (*resource.MetaValues, error)
	Next() bool
}

type Writer interface {
	WriteHeader() error
	WriteRow(interface{}) error
	Flush()
}
