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

	// add xray tracing debug info
	os.Setenv("AWS_XRAY_CONTEXT_MISSING", "LOG_ERROR")

	// retrieve sqs queue url from lambda variable
	urlqueue := os.Getenv("sqs_queue_url")

	// retrieve amount of messages to send
	mc := os.Getenv("total_message_count")

	// convert string to int
	msgc, _ := strconv.Atoi(mc)

	// divide the user submitted message count by 10 as SendMessageBatch is used
	msgcdiv := msgc / 10

	// print message with total amount of messages and queue url
	fmt.Println("sending " + strconv.Itoa(msgcdiv) + " messages to " + urlqueue)

	// setup a session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create a session with sqs, instrumented with xray
	svc := sqs.New(sess)
	xray.AWS(svc.Client)

	// run for every message group,
	for tot := 0; tot < (msgcdiv); tot++ {

		// retrieve context for xray and start subsegment for xray
		_, Seg := xray.BeginSubsegment(ctx, "sqs")

		// send the message to the sqs queue
		xray.Capture(ctx, "SendMsgBatch", func(ctx1 context.Context) error {
			entries := []*sqs.SendMessageBatchRequestEntry{}

			// generate 10 random message enrtries
			for i := 0; i == 10; i++ {

				// generate a random number of four digits
				msg := strconv.Itoa(rand.Intn(9999))

				// create the batch message request
				entry := sqs.SendMessageBatchRequestEntry{
					Id:          aws.String(strconv.Itoa(i)),
					MessageBody: aws.String(msg),
				}
				entries = append(entries, &entry)
			}

			// send the batch message request
			_, err := svc.SendMessageBatchWithContext(ctx, &sqs.SendMessageBatchInput{
				Entries:  entries,
				QueueUrl: aws.String(urlqueue),
			})

			// print an error if sending of messages failed
			if err != nil {
				fmt.Println(err)
			}

			return nil
		})

		// print total messages completed
		fmt.Println("sent " + strconv.Itoa((tot * 10)) + " messages")

		// close xray segments
		Seg.Close(nil)
	}
}

// start lambda handler
func main() {
	lambda.Start(handler)
}
