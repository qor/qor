package csv

func New(filename string) *CSV {
	return &CSV{Filename: filename}
}

type CSV struct {
	Filename string
	records  [][]string
}

func (csv *CSV) WriteLog(string) {
}
