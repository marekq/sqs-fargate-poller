package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	_ "github.com/aws/aws-xray-sdk-go/plugins/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/aws/aws-xray-sdk-go/xraylog"
)

// set the poll timeout value in seconds and the SQS queue URL
var (
	polltimeout = aws.Int64(20)
	sqsURL      = os.Getenv("sqs_queue_url")
	sqsName     = os.Getenv("SQS_NAME")
)

func init() {
	xray.SetLogger(xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelInfo))
	xray.Configure(xray.Config{})
	os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")
}

// retrieve a message from SQS
func pollmsg(svc *sqs.SQS) {
	ctx := context.Background()
	ctx, seg := xray.BeginSegment(ctx, "pollmsg")
	defer seg.Close(nil)

	result, _ := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(sqsURL),
		MaxNumberOfMessages:   aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{"All"}),
		WaitTimeSeconds:       polltimeout,
	})

	// check if the response contains messages
	if result.Messages != nil {
		msghandler := result.Messages[0].ReceiptHandle

		fmt.Println(result.Messages)
		xray.AddMetadata(ctx, "MsgReceived", result.Messages)

		// print the message and delete it from queue
		_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(sqsURL),
			ReceiptHandle: msghandler,
		})

		// print error if the message can't be deleted
		if err != nil {
			fmt.Println("message not deleted")
			xray.AddMetadata(ctx, "MsgDeleteFail", msghandler)
		} else {
			xray.AddMetadata(ctx, "MsgDeleteSuccess", msghandler)
		}

	} else {
		xray.AddMetadata(ctx, "MsgGetFailed", "message")
	}
}

// connect to SQS and run an infinite loop on pollers
//func handler(ctx context.Context) error {
func handler() {
	xray.Configure(xray.Config{
		DaemonAddr:     "127.0.0.1:2000",
		ServiceVersion: "1.2.3",
		LogLevel:       "debug",
	})

	xray.SetLogger(xraylog.NewDefaultLogger(os.Stderr, xraylog.LogLevelError))

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// start new session and instrument with XRay
	svc := sqs.New(sess)

	// print message queue url
	fmt.Println(" sqsName" + sqsName + "\n sqsURL " + sqsURL)

	// start infinite loop to poll queue
	for {
		pollmsg(svc)
	}

}

func main() {
	fmt.Println("starting container")
	handler()
}
