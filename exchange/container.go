package exchange

import "github.com/qor/qor/resource"

type Container interface {
	NewReader(*Resource) (Rows, error)
	NewWriter(*Resource) (Writer, error)
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
