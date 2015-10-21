package csv

import (
	"github.com/qor/qor/exchange"
	"github.com/qor/qor/resource"
)

func New(filename string) *CSV {
	return &CSV{Filename: filename}
}

type CSV struct {
	Filename string
}

func (csv *CSV) Rows() (exchange.Rows, error) {
	return Rows{}, nil
}

func (csv *CSV) WriteRow(*resource.MetaValues) {

}

func (csv *CSV) WriteLog(string) {
}

type Rows struct {
}

func (Rows) Columns() []string {
	return []string{}
}

func (Rows) CurrentColumn() (*resource.MetaValues, error) {
	return nil, nil
}

func (Rows) Next() bool {
	return false
}
