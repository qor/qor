package exchange

import "github.com/qor/qor/resource"

type Container interface {
	NewReader(*Resource) (Rows, error)
	NewWriter(*Resource) (Writer, error)
	WriteLog(string)
}

type Rows interface {
	Columns() []string
	CurrentColumn() (*resource.MetaValues, error)
	Next() bool
}

type Writer interface {
	WriterHeader() error
	WriteRow(*resource.MetaValues) error
}
