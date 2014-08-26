package exchange

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"code.google.com/p/google-api-go-client/drive/v2"
	"github.com/golang/oauth2"
	"github.com/golang/oauth2/google"
)

const (
	ExcelContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	CSVContentType   = "text/csv"
)

type googleDriveFile struct {
	*os.File
	size        int
	contentType string
}

func (f googleDriveFile) Len() int {
	return f.size
}

func (f googleDriveFile) ContentType() string {
	return f.contentType
}

type GoogleDriveConverter struct {
	config *oauth2.JWTConfig
}

// More explanations, see: http://godoc.org/github.com/golang/oauth2
//
//	 The contents of your RSA private key or your PEM file
//	 that contains a private key.
//	 If you have a p12 file instead, you
//	 can use `openssl` to export the private key into a pem file.
//
//	    $ openssl pkcs12 -in key.p12 -out key.pem -nodes
//
//	 It only supports PEM containers with no passphrase.
//
func NewGoogleDriveConverter(clientEmail, keyFilePath string) (gdc *GoogleDriveConverter, err error) {
	key, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		return
	}
	gdc = new(GoogleDriveConverter)
	gdc.config, err = google.NewServiceAccountConfig(&oauth2.JWTOptions{
		Email:      clientEmail,
		PrivateKey: key,
		Scopes:     []string{"https://www.googleapis.com/auth/drive"},
	})
	if err != nil {
		return
	}

	return
}

func (gdc *GoogleDriveConverter) Convert(path, from, to string) (r io.Reader, err error) {
	gdf := &googleDriveFile{contentType: from}
	gdf.File, err = os.Open(path)
	if err != nil {
		return
	}
	stat, err := gdf.File.Stat()
	if err != nil {
		return
	}
	gdf.size = int(stat.Size())

	client := http.Client{Transport: gdc.config.NewTransport()}
	svc, err := drive.New(&client)
	if err != nil {
		return
	}
	scheme := &drive.File{MimeType: from}
	file, err := svc.Files.Insert(scheme).Media(gdf).Convert(true).Do()
	if err != nil {
		return
	}

	csvLink, ok := file.ExportLinks[to]
	if !ok {
		err = fmt.Errorf("can't find ExportLinks for %s", to)
		return
	}

	resp, err := client.Get(csvLink)
	if err != nil {
		return
	}
	r = resp.Body

	return
}
