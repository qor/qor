package csv

import (
	"encoding/csv"
	"os"
	"strconv"

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
	var rows = Rows{CSV: c, Resource: res}

	csvfile, err := os.Open(c.Filename)
	if err == nil {
		defer csvfile.Close()
		reader := csv.NewReader(csvfile)
		reader.TrimLeadingSpace = true
		rows.records, err = reader.ReadAll()
		rows.total = len(rows.records)
		if res.Config.WithoutHeader {
			rows.current = -1
		}
	}

	return &rows, err
}

func (csv *CSV) WriteRow(*resource.MetaValues) {
}

func (csv *CSV) WriteLog(string) {
}

type Rows struct {
	*CSV
	Resource *exchange.Resource
	current  int
	total    int
}

func (rows Rows) Columns() (results []string) {
	if rows.total > 0 {
		if rows.Resource.Config.WithoutHeader {
			for i := 0; i <= len(rows.records[0]); i++ {
				results = append(results, strconv.Itoa(i))
			}
		} else {
			return rows.records[0]
		}
	}
	return
}

func (rows Rows) CurrentColumn() (*resource.MetaValues, error) {
	var metaValues resource.MetaValues
	columns := rows.Columns()

	for index, column := range columns {
		metaValue := resource.MetaValue{
			Name:  column,
			Value: rows.records[rows.current][index],
		}
		if meta := rows.Resource.GetMeta(column); meta != nil {
			metaValue.Meta = meta
		}
		metaValues.Values = append(metaValues.Values, &metaValue)
	}

	return &metaValues, nil
}

func (rows *Rows) Next() bool {
	if rows.total > rows.current+1 {
		rows.current += 1
		return true
	}
	return false
}
