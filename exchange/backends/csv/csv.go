package csv

import (
	"encoding/csv"
	"errors"
	"os"

	"github.com/qor/qor/exchange"
	"github.com/qor/qor/resource"
)

func New(filename string) *CSV {
	return &CSV{Filename: filename}
}

type CSV struct {
	Filename string
	records  [][]string
}

func (c *CSV) Rows(res *exchange.Resource) (exchange.Rows, error) {
	var rows = Rows{CSV: c}

	csvfile, err := os.Open(c.Filename)
	if err == nil {
		defer csvfile.Close()
		reader := csv.NewReader(csvfile)
		rows.records, err = reader.ReadAll()
		rows.total = len(rows.records)
	}

	return &rows, err
}

func (csv *CSV) WriteRow(*resource.MetaValues) {
}

func (csv *CSV) WriteLog(string) {
}

type Rows struct {
	*CSV
	current int
	total   int
}

func (rows Rows) Columns() []string {
	if rows.total > 0 {
		return rows.records[0]
	}
	return []string{}
}

func (Rows) CurrentColumn() (*resource.MetaValues, error) {
	return nil, errors.New("not implemented")
}

func (rows *Rows) Next() bool {
	if rows.total > rows.current {
		rows.current += 1
		return true
	}
	return false
}
