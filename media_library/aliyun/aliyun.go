package aliyun

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/qor/qor/media_library"
	"github.com/qor/qor/media_library/aliyun/config"
	"github.com/sunfmin/ali-oss"
	"log"
)

type Aliyun struct {
	media_library.Base
}

var aliossClient *alioss.Client

func init() {
	aliossClient = alioss.NewClient(config.AliOSSAccessKey, config.AliOSSAccessSecret)
}

func getBucket(option *media_library.Option) string {
	if bucket := option.Get("bucket"); bucket != "" {
		return bucket
	}

	return config.AliOSSBucket
}

func getEndpoint(option *media_library.Option) string {
	if endpoint := option.Get("endpoint"); endpoint != "" {
		return endpoint
	}
	return getBucket(option) + "." + config.AliOSSRegion + "." + "aliyuncs.com"
}

func (s Aliyun) GetURLTemplate(option *media_library.Option) (path string) {
	if path = option.Get("URL"); path == "" {
		path = "/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	}

	path = "//" + getEndpoint(option) + path

	log.Printf("path = %+v\n", path)
	return
}

func (s Aliyun) Store(url string, option *media_library.Option, reader io.Reader) (err error) {

	bucket := alioss.NewBucket(config.AliOSSBucket, alioss.BucketRegion(config.AliOSSRegion), aliossClient)

	err = bucket.Put(url, reader, nil)

	if err != nil {
		return
	}

	return
}

func (s Aliyun) Retrieve(url string) (*os.File, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if file, err := ioutil.TempFile("/tmp", "aliyun"); err == nil {
		_, err := io.Copy(file, response.Body)
		return file, err
	} else {
		return nil, err
	}
}
