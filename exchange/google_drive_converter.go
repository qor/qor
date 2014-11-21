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

type ContentType string

const (
	ExcelContentType ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	CSVContentType   ContentType = "text/csv"
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
	flow *oauth2.Flow
}

// How to create a google api service account: https://developers.google.com/drive/web/service-accounts
// About google-api-go-client: https://code.google.com/p/google-api-go-client/wiki/GettingStarted
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
//	More explanations, see: http://godoc.org/github.com/golang/oauth2
func NewGoogleDriveConverter(clientEmail, keyFilePath string) (gdc *GoogleDriveConverter, err error) {
	key, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		return
	}
	gdc = new(GoogleDriveConverter)
	gdc.flow, err = oauth2.New(
		oauth2.JWTClient(clientEmail, key),
		oauth2.Scope("https://www.googleapis.com/auth/drive"),
		google.JWTEndpoint(),
		// oauth2.Subject("user@example.com"),
	)
	if err != nil {
		return
	}

	return
}

// NewGoogleDriveConverterByJSONKey will accept a json key file downloaded from google
// project and return a GoogleDriveConverter. Built for convinence.
func NewGoogleDriveConverterByJSONKey(filename string) (gdc *GoogleDriveConverter, err error) {
	// rawkey, err := ioutil.ReadFile(jsonKey)
	// if err != nil {
	// 	return
	// }
	// key := map[string]string{}
	// err = json.Unmarshal(rawkey, &key)
	// if err != nil {
	// 	return
	// }

	gdc = new(GoogleDriveConverter)
	gdc.flow, err = oauth2.New(
		google.ServiceAccountJSONKey(filename),
		oauth2.Scope("https://www.googleapis.com/auth/drive"),
	)
	// gdc.config, err = google.NewServiceAccountConfig(&oauth2.JWTOptions{
	// 	Email:      key["client_email"],
	// 	PrivateKey: []byte(key["private_key"]),
	// 	Scopes:     []string{"https://www.googleapis.com/auth/drive"},
	// })
	if err != nil {
		return
	}

	return
}

func (gdc *GoogleDriveConverter) Convert(path string, from, to ContentType) (r io.Reader, err error) {
	gdf := &googleDriveFile{contentType: string(from)}
	gdf.File, err = os.Open(path)
	if err != nil {
		return
	}
	stat, err := gdf.File.Stat()
	if err != nil {
		return
	}
	gdf.size = int(stat.Size())

	client := http.Client{Transport: gdc.flow.NewTransport()}
	svc, err := drive.New(&client)
	if err != nil {
		return
	}
	scheme := &drive.File{MimeType: string(from)}
	file, err := svc.Files.Insert(scheme).Media(gdf).Convert(true).Do()
	if err != nil {
		return
	}

	csvLink, ok := file.ExportLinks[string(to)]
	if !ok {
		err = fmt.Errorf("can't find ExportLinks for %s", string(to))
		return
	}

	resp, err := client.Get(csvLink)
	if err != nil {
		return
	}
	r = resp.Body

	return
}
