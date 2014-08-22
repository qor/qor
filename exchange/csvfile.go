package exchange

import (
	"encoding/csv"

	"io"
)

type CSVFile struct {
	lines [][]string
}

func NewCSVFile(r io.Reader) (f *CSVFile, err error) {
	f = new(CSVFile)
	f.lines, err = csv.NewReader(r).ReadAll()
	return
}

func (f *CSVFile) TotalLines() (num int) {
	return len(f.lines)
}

func (f *CSVFile) Line(l int) []string {
	return f[l]
}
