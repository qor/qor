package csv

import (
	"encoding/csv"
	"os"

	"github.com/qor/qor/exchange"
	"github.com/qor/qor/resource"
)

func (c *CSV) NewWriter(res *exchange.Resource) (exchange.Writer, error) {
	writer := &Writer{CSV: c, Resource: res, metas: res.GetMetas([]string{})}

	csvfile, err := os.OpenFile(c.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err == nil {
		writer.Writer = csv.NewWriter(csvfile)
	}

	return writer, err
}

type Writer struct {
	*CSV
	Resource *exchange.Resource
	Writer   *csv.Writer
	metas    []resource.Metaor
}

func (writer *Writer) WriteHeader() error {
	if !writer.Resource.Config.WithoutHeader {
		var results []string
		for _, meta := range writer.metas {
			results = append(results, meta.GetName())
		}
		writer.Writer.Write(results)
	}
	return nil
}

func (writer *Writer) WriteRow(record interface{}) error {
	for _, meta := range writer.metas {
		meta.GetName()
	}
	return nil
}

func (writer *Writer) Flush() {
	writer.Writer.Flush()
}
