package S3

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/qor/qor/media_library"
)

type S3 struct {
	media_library.Base
}

var S3Client *s3.S3

func Init(bucket string, config *aws.Config) {
	S3Client = s3.New(config)
}

func getBucket(option *media_library.Option) string {
	if bucket := os.Getenv("S3Bucket"); bucket != "" {
		return bucket
	}
	return option.Get("bucket")
}

func getEndpoint(option *media_library.Option) string {
	if endpoint := option.Get("endpoint"); endpoint != "" {
		return endpoint
	}
	return getBucket(option) + "." + S3Client.Config.Endpoint
}

func (s S3) GetURLTemplate(option *media_library.Option) (path string) {
	if path = option.Get("URL"); path == "" {
		path = "//" + getEndpoint(option) + "/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	}

	return "//" + getEndpoint(option) + path
}

func (s S3) Store(url string, option *media_library.Option, reader io.Reader) error {
	var buffer = []byte{}
	reader.Read(buffer)
	fileBytes := bytes.NewReader(buffer)

	params := &s3.PutObjectInput{
		Bucket:        aws.String(getBucket(option)), // required
		Key:           aws.String(url),               // required
		ACL:           aws.String("public-read"),
		Body:          fileBytes,
		ContentLength: aws.Long(int64(fileBytes.Len())),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"), //required
		},
	}
	// see more at http://godoc.org/github.com/aws/aws-sdk-go/service/s3#S3.PutObject

	_, err := S3Client.PutObject(params)
	return err
}

func (s S3) Retrieve(url string) (*os.File, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if file, err := ioutil.TempFile("/tmp", "s3"); err == nil {
		_, err := io.Copy(file, response.Body)
		return file, err
	} else {
		return nil, err
	}
}
