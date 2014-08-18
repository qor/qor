package exchange

type CSVFile struct {
}

func NewCSVFile() *CSVFile {
	f := new(CSVFile)
	return f
}

func (x *CSVFile) TotalLines() (num int) {
	return
}

func (x *CSVFile) Line(l int) (fields []string) {
	return
}
