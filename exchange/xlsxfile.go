package exchange

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/tealeg/xlsx"
)

type XLSXFile struct {
	lines [][]string
}

func NewXLSXFile(filename string) (nxf *XLSXFile, err error) {
	zr, err := zip.OpenReader(filename)
	if err != nil {
		return
	}

	return NewXLSXFileZipReader(zr)
}

func NewXLSXFileReader(r io.Reader) (nxf *XLSXFile, err error) {
	f, err := ioutil.TempFile("", "qor.exchange.")
	if err != nil {
		return
	}
	defer func() { f.Close() }()
	_, err = io.Copy(f, r)
	if err != nil {
		return
	}
	defer func() { os.Remove(f.Name()) }()

	return NewXLSXFile(f.Name())
}

func NewXLSXFileZipReader(zr *zip.ReadCloser) (nxf *XLSXFile, err error) {
	xf, err := xlsx.ReadZip(zr)
	if err != nil {
		return
	}

	if len(xf.Sheets) == 0 {
		err = errors.New("exchange: find no sheets in file")
		return
	}

	// TODO: support multiple sheets in future
	sheet := xf.Sheets[0]

	if len(sheet.Rows) == 0 {
		return
	}

	nxf = new(XLSXFile)
	for _, row := range sheet.Rows {
		if len(row.Cells) == 0 {
			continue
		}

		var lines []string
		skipline := true
		for _, cell := range row.Cells {
			field := strings.TrimSpace(cell.Value)
			if skipline {
				if field == "" {
					continue
				}
			}

			lines = append(lines, field)
			skipline = false
		}

		if skipline {
			continue
		}

		nxf.lines = append(nxf.lines, lines)
	}

	return
}

func preprocessXLSXFile(xf *xlsx.File) (totalLines int, nxf *xlsx.File) {
	nxf = new(xlsx.File)

	return
}

func (x *XLSXFile) TotalLines() (num int) {
	return len(x.lines)
}

func (x *XLSXFile) Line(l int) (fields []string) {
	return x.lines[l]
}
