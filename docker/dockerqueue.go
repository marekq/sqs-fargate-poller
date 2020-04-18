package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// set the poll timeout value in seconds and the SQS queue URL
var (
	polltimeout = aws.Int64(20)
	sqsURL      = os.Getenv("sqs_queue_url")
)

// connect to SQS and run an infinite loop on pollers
func main() {

	//ecs.Init()

	xray.Configure(xray.Config{
		DaemonAddr:     "127.0.0.1:2000",
		ServiceVersion: "1.2.3",
		LogLevel:       "trace",
	})

	os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")

	// print message queue url
	fmt.Println("sqsURL " + sqsURL)

	// open a new session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// connect to sqs
	svc := sqs.New(sess)
	xray.AWS(svc.Client)
	c := 0

	// start infinite loop to poll queue
	for {

		// create an xray segement
		ctx, Seg := xray.BeginSegment(context.Background(), "sqs")
		_, SubSeg := xray.BeginSubsegment(ctx, "subseg")

		defer SubSeg.Close(nil)
		defer Seg.Close(nil)
		ctx := context.Background()

		// capture the receive message request with xray
		xray.Capture(ctx, "ReceiveMsg", func(ctx1 context.Context) error {
			result, _ := svc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
				QueueUrl:              aws.String(sqsURL),
				MaxNumberOfMessages:   aws.Int64(1),
				MessageAttributeNames: aws.StringSlice([]string{"All"}),
				WaitTimeSeconds:       polltimeout,
			})

			if len(result.Messages) > 0 {
				msghandler := result.Messages[0].ReceiptHandle
				body := result.Messages[0].Body

				svc.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(sqsURL),
					ReceiptHandle: msghandler,
				})
				fmt.Println(strconv.Itoa(c) + " received and deleting " + *body)
			} else {
				fmt.Println("no messages retrieved")
			}

			// increase total count
			c++

			return nil
		})
	}
}
