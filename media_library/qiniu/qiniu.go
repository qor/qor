package qiniu

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	qnconf "github.com/qiniu/api.v6/conf"
	qnio "github.com/qiniu/api.v6/io"
	qnrs "github.com/qiniu/api.v6/rs"
	"github.com/qor/qor/media_library"
)

var (
	Uptoken       = ""
	DownloadToken = ""
	Bucket        = ""
	EndPoint      = ""
)

type Qiniu struct {
	media_library.Base
}

func init() {
	Bucket = os.Getenv("QOR_QINIU_BUCKET")
	qnconf.ACCESS_KEY = os.Getenv("QOR_QINIU_ACCESS_KEY")
	qnconf.SECRET_KEY = os.Getenv("QOR_QINIU_SECRET_KEY")
	EndPoint = os.Getenv("QOR_QINIU_ENDPOINT")
	Uptoken = uptoken(Bucket)
}

func uptoken(bucketName string) string {
	putPolicy := qnrs.PutPolicy{
		Scope: bucketName,
		//CallbackUrl: callbackUrl,
		//CallbackBody:callbackBody,
		//ReturnUrl:   returnUrl,
		//ReturnBody:  returnBody,
		//AsyncOps:    asyncOps,
		//EndUser:     endUser,
		//Expires:     expires,
	}
	return putPolicy.Token(nil)
}

func getBucket(option *media_library.Option) string {
	if bucket := option.Get("bucket"); bucket != "" {
		return bucket
	}

	return Bucket
}

func getEndpoint(option *media_library.Option) string {
	if endpoint := option.Get("endpoint"); endpoint != "" {
		return endpoint
	}
	return EndPoint + "/@"
}

func (q Qiniu) GetURLTemplate(option *media_library.Option) (path string) {
	if path = option.Get("URL"); path == "" {
		path = "/{{class}}/{{primary_key}}/{{column}}/{{filename_with_hash}}"
	}

	path = "//" + getEndpoint(option) + path
	return
}

func (q Qiniu) Store(url string, option *media_library.Option, reader io.Reader) (err error) {

	var ret qnio.PutRet
	path := strings.Replace(url, "//"+getEndpoint(option), "", -1)
	// ret       变量用于存取返回的信息，详情见 io.PutRet
	// uptoken   为业务服务器端生成的上传口令
	// r         为io.Reader类型，用于从其读取数据
	// extra     为上传文件的额外信息,可为空， 详情见 io.PutExtra, 可选
	err = qnio.Put(nil, &ret, Uptoken, path, reader, nil)

	if err != nil {
		//上传产生错误
		log.Print("io.Put failed:", err)
		return
	}

	//上传成功，处理返回值
	// log.Print(ret.Hash, ret.Key)
	return
}

func (q Qiniu) Retrieve(url string) (*os.File, error) {
	response, err := http.Get("http:" + url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if file, err := ioutil.TempFile("/tmp", "qiniu"); err == nil {
		_, err := io.Copy(file, response.Body)
		return file, err
	} else {
		return nil, err
	}
}
