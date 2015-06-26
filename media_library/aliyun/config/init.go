package config

import (
	"log"
	"os"
	"strings"
)

// this is configured from env variables
var (
	AliOSSAccessKey    string
	AliOSSAccessSecret string
	AliOSSBucket       string
	AliOSSRegion       string
	AliOSSEndpoint     string
)

func init() {
	AliOSSAccessKey = envOrPanic("QOR_ALIOSS_ACCESS_KEY", false)
	AliOSSAccessSecret = envOrPanic("QOR_ALIOSS_ACCESS_SECRET", false)
	AliOSSBucket = envOrPanic("QOR_ALIOSS_BUCKET", false)
	AliOSSRegion = envOrPanic("QOR_ALIOSS_REGION", false)
	AliOSSEndpoint = envOrPanic("QOR_ALIOSS_ENDPOINT", false)
}

func envOrPanic(key string, allowEmpty bool) (r string) {
	r = os.Getenv(key)
	if r == "" && !allowEmpty {
		panic("env " + key + " is not set")
	}
	logValue := r
	if strings.Contains(logValue, "PASSWORD") || strings.Contains(logValue, "SECRET") {
		logValue = "<HIDDEN>"
	}
	log.Printf("Configure: %s = %s\n", key, logValue)
	return
}
