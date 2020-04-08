package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-xray-sdk-go/awsplugins/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// set the poll timeout value in seconds and the SQS queue URL
var (
	polltimeout = aws.Int64(20)
	sqsURL      = os.Getenv("sqs_queue_url")
	sqsName     = os.Getenv("SQS_NAME")
)

func init() {
	os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")
	ecs.Init()

	xray.Configure(xray.Config{
		DaemonAddr:     "127.0.0.1:2000",
		ServiceVersion: "1.2.3",
		LogLevel:       "trace",
	})
}

// connect to SQS and run an infinite loop on pollers
func main() {
	// print message queue url
	fmt.Println(" sqsName" + sqsName + "\n sqsURL " + sqsURL)

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
		ctx, Seg := xray.BeginSegment(context.Background(), "sqs")
		_, SubSeg := xray.BeginSubsegment(ctx, "subsegment-name")

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
				fmt.Println("received and deleting " + *body)
				xray.AddMetadata(ctx, "SuccessMsg", "0")

			} else {
				xray.AddMetadata(ctx, "FailMsg", "0")
			}

			return nil
		})

		SubSeg.Close(nil)
		Seg.Close(nil)

		c++
		fmt.Println(c)
	}
}
