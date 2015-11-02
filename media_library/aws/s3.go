package aws

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/qor/qor/media_library"
)

type S3 struct {
	media_library.Base
}

var S3Client *s3.S3
var AwsRegion = os.Getenv("QOR_AWS_REGION")
var AwsAccessKeyId = os.Getenv("QOR_AWS_ACCESS_KEY_ID")
var AwsSecretAccessKey = os.Getenv("QOR_AWS_SECRET_ACCESS_KEY")
var AwsSessionToken = os.Getenv("QOR_AWS_SESSION_TOKEN")

func s3client() *s3.S3 {
	if S3Client == nil {
		creds := credentials.NewStaticCredentials(AwsAccessKeyId, AwsSecretAccessKey, AwsSessionToken)

		if _, err := creds.Get(); err == nil {
			S3Client = s3.New(session.New(), &aws.Config{
				Region:      &AwsRegion,
				Credentials: creds,
			})
		}
	}
	return S3Client
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

	return getBucket(option) + "." + *S3Client.Config.Endpoint
}

func (s S3) GetURLTemplate(option *media_library.Option) (path string) {
	if path = option.Get("URL"); path == "" {
		path = "/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	}

	return "//" + getEndpoint(option) + path
}

func (s S3) Store(url string, option *media_library.Option, reader io.Reader) error {
	var buffer = []byte{}
	reader.Read(buffer)
	fileBytes := bytes.NewReader(buffer)

	path := strings.Replace(url, "//"+getEndpoint(option), "", -1)
	fileBytesLen := int64(fileBytes.Len())

	params := &s3.PutObjectInput{
		Bucket:        aws.String(getBucket(option)), // required
		Key:           aws.String(path),              // required
		ACL:           aws.String("public-read"),
		Body:          fileBytes,
		ContentLength: &fileBytesLen,
		ContentType:   aws.String(http.DetectContentType(buffer)),
		Metadata: map[string]*string{
			"Key": aws.String("MetadataValue"), //required
		},
	}

	_, err := S3Client.PutObject(params)
	return err
}

func (s S3) Retrieve(url string) (*os.File, error) {
	response, err := http.Get("http:" + url)
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
