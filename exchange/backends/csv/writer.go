package csv

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/qor/qor"
	"github.com/qor/qor/exchange"
	"github.com/qor/qor/resource"
	"github.com/qor/qor/roles"
)

func (c *CSV) NewWriter(res *exchange.Resource, context *qor.Context) (exchange.Writer, error) {
	writer := &Writer{CSV: c, Resource: res, context: context}

	var metas []resource.Metaor
	for _, meta := range res.GetMetas([]string{}) {
		if meta.HasPermission(roles.Read, context) {
			metas = append(metas, meta)
		}
	}
	writer.metas = metas

	csvfile, err := os.OpenFile(c.Filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err == nil {
		writer.Writer = csv.NewWriter(csvfile)
	}

	return writer, err
}

type Writer struct {
	*CSV
	context  *qor.Context
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
	var results []string
	for _, meta := range writer.metas {
		results = append(results, fmt.Sprint(meta.GetValuer()(record, writer.context)))
	}
	writer.Writer.Write(results)
	return nil
}

func (writer *Writer) Flush() {
	writer.Writer.Flush()
}
