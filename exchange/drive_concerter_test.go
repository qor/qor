package exchange

import (
	"flag"
	"io"
	"os"

	"testing"
)

var googleAPIEmail, googleAPIKeyFile, jsonKey, excelfile, csvfile string

func init() {
	flag.StringVar(&googleAPIEmail, "googleAPIEmail", "", "")
	flag.StringVar(&googleAPIKeyFile, "googleAPIKeyFile", "", "")
	flag.StringVar(&jsonKey, "jsonKey", "", "")
	flag.StringVar(&excelfile, "excelfile", "", "")
	flag.StringVar(&csvfile, "csvfile", "", "")
}

func TestNewGoogleDriveCSVConverter(t *testing.T) {
	flag.Parse()

	var err error
	var converter *GoogleDriveConverter
	if googleAPIEmail != "" && googleAPIKeyFile != "" {
		converter, err = NewGoogleDriveConverter(googleAPIEmail, googleAPIKeyFile)
		if err != nil {
			t.Fatal(err)
		}
	} else if jsonKey != "" {
		converter, err = NewGoogleDriveConverterByJSONKey(jsonKey)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		return
	}

	if excelfile != "" {
		csvsrc, err := converter.Convert(excelfile, ExcelContentType, CSVContentType)
		if err != nil {
			t.Fatalf("Can't convert excel to csv: %s", err)
		}
		csvdst, err := os.OpenFile("test.csv", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(csvdst, csvsrc)
		if err != nil {
			t.Error(err)
		}
	}

	if csvfile != "" {
		excelsrc, err := converter.Convert(csvfile, CSVContentType, ExcelContentType)
		if err != nil {
			t.Fatalf("Can't convert csv to excel: %s", err)
		}
		exceldst, err := os.OpenFile("test.xlsx", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(exceldst, excelsrc)
		if err != nil {
			t.Error(err)
		}
	}
}
