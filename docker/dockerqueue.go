package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-xray-sdk-go/awsplugins/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/aws/aws-xray-sdk-go/xraylog"
)

var (
	// set the sqs polling timeout to 20 seconds
	polltimeout = aws.Int64(20)

	// retrieve the sqs queue url from the environment variable
	sqsURL = os.Getenv("sqs_queue_url")

	// set a count value to keep track of total amount of messages processed
	c = 1
)

// retrieve a message from SQS
func pollmsg(svc *sqs.SQS) {
	ctx := context.Background()
	ctx, Seg := xray.BeginSegment(ctx, "readmsg")
	_, SubSeg := xray.BeginSubsegment(ctx, "subseg")

	// capture the request and response times to sqs using xray
	xray.Capture(ctx, "readmsg", func(ctx1 context.Context) error {
		result, _ := svc.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:              aws.String(sqsURL),
			MaxNumberOfMessages:   aws.Int64(1),
			MessageAttributeNames: aws.StringSlice([]string{"All"}),
			WaitTimeSeconds:       polltimeout,
		})

		// check if any responses were retrieved in the request
		if result.Messages != nil {
			xray.AddMetadata(ctx, "MsgReceived", result.Messages)
			msghandler := result.Messages[0].ReceiptHandle

			// print the message and delete it from queue
			_, err := svc.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(sqsURL),
				ReceiptHandle: msghandler,
			})

			// print error if the message can't be deleted
			if err != nil {
				fmt.Println("message not deleted")
				xray.AddMetadata(ctx, "MsgDeleteFail", msghandler)

				// print the succesful outcome of the retrieve
			} else {
				c++
				fmt.Println(strconv.Itoa(c) + " messages processed")
				xray.AddMetadata(ctx, "MsgDeleteSuccess", msghandler)
			}

			// print message if no messages were retrieved during long polling
		} else {
			fmt.Println("no messages processed in 20s window")
			xray.AddMetadata(ctx, "MsgGetFailed", "message")
		}

		return nil
	})
	SubSeg.Close(nil)
	Seg.Close(nil)
}

// connect to SQS and run an infinite loop on pollers
func main() {

	// TODO temporary workaround - wait 10 seconds to prevent the xray sidecar is up and running
	// it appears the go routine can get stuck if it cannot ship xray traces directly
	time.Sleep(10 * time.Second)

	// init ECS plugin
	ecs.Init()

	// configure xray to send metrics to the sidecar
	xray.Configure(xray.Config{
		DaemonAddr:     "127.0.0.1:2000",
		ServiceVersion: "1.2.3",
		LogLevel:       "trace",
	})

	// enable error logging
	xray.SetLogger(xraylog.NewDefaultLogger(os.Stdout, xraylog.LogLevelInfo))

	// create a session with sqs
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// start new session and instrument with XRay
	svc := sqs.New(sess)
	xray.AWS(svc.Client)

	// print message queue url
	fmt.Println("sqsURL " + sqsURL)

	// start infinite loop to poll queue
	for {
		pollmsg(svc)
	}
}
