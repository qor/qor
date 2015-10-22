package csv

import (
	"github.com/qor/qor/exchange"
	"github.com/qor/qor/resource"
)

func (c *CSV) NewWriter(res *exchange.Resource) (exchange.Writer, error) {
	writer := &Writer{}
	return writer, nil
}

type Writer struct {
}

func (*Writer) WriterHeader() error {
	return nil
}

func (*Writer) WriteRow(*resource.MetaValues) error {
	return nil
}
