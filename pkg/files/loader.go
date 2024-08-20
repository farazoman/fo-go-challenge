package files

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Loader interface {
	Load(path string) (string, error)
}

type S3SingleBucketLoader struct {
	Bucket string
}

func (loader S3SingleBucketLoader) Load(objectKey string) (string, error) {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	client := s3.NewFromConfig(cfg)

	objectKey = strings.TrimSpace(objectKey)

	fmt.Printf("DEBUG: bucket: %s, key: %s\n", loader.Bucket, objectKey)

	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(loader.Bucket),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, resp.Body); err != nil {
		log.Fatal(err)
	}

	content := buf.String()
	fmt.Printf("DEBUG: S3 File Content is\n%s\n", content)
	return content, nil
}

var _ Loader = S3SingleBucketLoader{}

type SystemLoader struct{}

func (loader SystemLoader) Load(path string) (string, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	fmt.Printf("DEBUG: %s", content)
	return string(content), nil
}

var _ Loader = SystemLoader{}
