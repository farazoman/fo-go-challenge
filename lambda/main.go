package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/farazoman/fo-go-challenge/pkg/files"
	"github.com/farazoman/fo-go-challenge/pkg/transactions"
)

func HandleRequest(ctx context.Context, event *events.S3Event) error {
	if event == nil {
		return fmt.Errorf("received nil event, %v", event)
	}

	fmt.Printf("Initilaizing reporter with loader in region %s\n", os.Getenv("AWS_REGION"))
	r := transactions.Reporter{
		Loader: files.S3SingleBucketLoader{
			Bucket: os.Getenv("NOTIFY_BUCKET_NAME"),
		},
	}

	fmt.Println("Iterating through records in event")
	for _, record := range event.Records {
		filePath := record.S3.Object.URLDecodedKey
		err := transactions.Process(filePath, r)
		if err != nil {
			log.Fatalf("Processing transction for file %s failed with error: \n%s", record.S3.Object.Key, err)
		}
	}

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
