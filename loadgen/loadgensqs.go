package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// main handler
func handler(ctx context.Context) {

	// add xray debug info
	os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")

	// retrieve sqs queue url from lambda variable
	urlqueue := os.Getenv("sqs_queue_url")

	// retrieve amount of messages to send
	mc := os.Getenv("total_message_count")

	// convert string to int
	msgc, _ := strconv.Atoi(mc)

	// print message with to be sent amount of messages and sqs queue url
	fmt.Println("start sending " + mc + " messages to " + urlqueue)

	// setup a session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create a session with sqs and instrument it with xray tracing
	svc := sqs.New(sess)
	xray.AWS(svc.Client)

	// create trace for every message group
	for tot := 0; tot < (msgc); tot++ {

		// retrieve context for xray and start subsegment
		_, Seg := xray.BeginSubsegment(ctx, "sqs")

		// send the message to the sqs queue
		xray.Capture(ctx, "SendMsg", func(ctx1 context.Context) error {
			ri := strconv.Itoa(rand.Intn(999))

			// send one message
			_, err := svc.SendMessageWithContext(ctx, &sqs.SendMessageInput{
				MessageBody: aws.String(ri),
				QueueUrl:    aws.String(urlqueue),
			})

			// print an error if message sending failed
			if err != nil {
				fmt.Println(err)
			}

			return nil
		})

		// print total messages completed
		fmt.Println("sent " + strconv.Itoa(tot) + " messages to queue")

		// close xray subsegments
		Seg.Close(nil)
	}
}

// start lambda handler
func main() {
	lambda.Start(handler)
}
