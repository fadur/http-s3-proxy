package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type config struct {
	endpoint  string
	region    string
	bucket    string
	keyPrefix string
}

var cfg config

func main() {
	cfg = config{
		endpoint:  os.Getenv("AWS_S3_ENDPOINT"),
		region:    os.Getenv("AWS_DEFAULT_REGION"),
		bucket:    os.Getenv("AWS_S3_BUCKET"),
		keyPrefix: os.Getenv("AWS_S3_KEY_PREFIX"),
	}
	http.HandleFunc("/", serveS3)
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	http.ListenAndServe(":80", nil)
}

func serveS3(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	sess := session.New(aws.NewConfig().WithRegion(cfg.region).WithEndpoint(cfg.endpoint).WithS3ForcePathStyle(true))
	req := &s3.GetObjectInput{
		Bucket: aws.String(cfg.bucket),
		Key:    aws.String(cfg.keyPrefix + path),
	}
	obj, err := s3.New(sess).GetObject(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	setStrHeader(w, "Cache-Control", obj.CacheControl)
	setStrHeader(w, "Expires", obj.Expires)
	setStrHeader(w, "Content-Disposition", obj.ContentDisposition)
	setStrHeader(w, "Content-Encoding", obj.ContentEncoding)
	setStrHeader(w, "Content-Language", obj.ContentLanguage)
	setIntHeader(w, "Content-Length", obj.ContentLength)
	setStrHeader(w, "Content-Range", obj.ContentRange)
	//setStrHeader(w, "Content-Type", obj.ContentType)
	setTimeHeader(w, "Last-Modified", obj.LastModified)
	io.Copy(w, obj.Body)
}

func setStrHeader(w http.ResponseWriter, key string, value *string) {
	if value != nil && len(*value) > 0 {
		w.Header().Add(key, *value)
	}
}

func setIntHeader(w http.ResponseWriter, key string, value *int64) {
	if value != nil && *value > 0 {
		w.Header().Add(key, strconv.FormatInt(*value, 10))
	}
}

func setTimeHeader(w http.ResponseWriter, key string, value *time.Time) {
	if value != nil && !reflect.DeepEqual(*value, time.Time{}) {
		w.Header().Add(key, value.UTC().Format(http.TimeFormat))
	}
}
