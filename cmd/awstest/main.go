package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/leetcode-golang-classroom/golang-async-api/internal/pkg/config"
)

func main() {
	// test list bucket
	ctx := context.Background()
	sdkConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you setup your AWS account?")
		return
	}
	appconfig := config.AppConfig
	s3Client := s3.NewFromConfig(sdkConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(appconfig.S3LocalstackEndPoint)
		options.UsePathStyle = true
	})
	out, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, bucket := range out.Buckets {
		fmt.Println(*bucket.Name)
	}
	// list queue urls
	sqsClient := sqs.NewFromConfig(sdkConfig, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(appconfig.LocalstackEndPoint)
	})
	sqsClientOutput, err := sqsClient.ListQueues(ctx, &sqs.ListQueuesInput{})
	if err != nil {
		log.Fatal(err)
	}
	for _, q := range sqsClientOutput.QueueUrls {
		fmt.Println(q)
	}
}
