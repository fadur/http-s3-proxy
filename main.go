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
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type config struct {
	endpoint  string
	region    string
	bucket    string
	keyPrefix string
	appPort   string
}

var cfg config

func main() {
	cfg = config{
		endpoint:  os.Getenv("AWS_S3_ENDPOINT"),
		region:    os.Getenv("AWS_DEFAULT_REGION"),
		bucket:    os.Getenv("AWS_S3_BUCKET"),
		keyPrefix: os.Getenv("AWS_S3_KEY_PREFIX"),
		appPort:   os.Getenv("APP_PORT"),
	}
	if cfg.appPort == "" {
		cfg.appPort = "80"
	}
	http.HandleFunc("/", serveS3)
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello")
	})
	fmt.Printf("http-s3-proxy listening on port %s\n", cfg.appPort)
	err := http.ListenAndServe(":"+cfg.appPort, nil)
	if err != nil {
		panic(err)
	}
}

func serveS3(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	sess := session.New(aws.NewConfig().WithRegion(cfg.region).WithEndpoint(cfg.endpoint).WithS3ForcePathStyle(true))
	ifModifiedSince, _ := http.ParseTime(r.Header.Get("If-Modified-Since"))
	req := &s3.GetObjectInput{
		Bucket:          aws.String(cfg.bucket),
		Key:             aws.String(cfg.keyPrefix + path),
		IfModifiedSince: &ifModifiedSince,
	}
	status := 200
	obj, err := s3.New(sess).GetObject(req)
	if err != nil {
		isOk := false
		if reqerr, ok := err.(awserr.RequestFailure); ok {
			status = reqerr.StatusCode()
			if reqerr.StatusCode() == 304 {
				isOk = true
			}
		} else {
			status = http.StatusInternalServerError
		}
		if !isOk {
			fmt.Printf("Error for '%s': %d, %s\n", path, status, err.Error())
			http.Error(w, err.Error(), status)
			return
		}
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
	w.WriteHeader(status)
	if status == 200 {
		io.Copy(w, obj.Body)
	}
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
