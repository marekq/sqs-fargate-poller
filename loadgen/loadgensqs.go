package main

import (
	"context"
	"fmt"
	"log"
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
	q := os.Getenv("sqs_queue_url")

	// retrieve amount of messages to send
	b := os.Getenv("total_message_count")
	c, _ := strconv.Atoi(b)

	// run goroutines in groups of 10
	d := 10

	// setup session
	fmt.Println("sending messages to " + q)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create a session with sqs, instrumented with xray
	svc := sqs.New(sess)
	xray.AWS(svc.Client)

	// keep track of total sent messages
	tot := 0

	// run for every message group
	for a := 0; a < (c / 10); a++ {

		// spawn go routines depending on total count
		for b := 0; b < d; b++ {

			// retrieve context for xray and start subsegment for xray
			_, Seg := xray.BeginSubsegment(ctx, "sqs")

			// generate a random number
			ri := strconv.Itoa(rand.Intn(9999))

			// send the message to the sqs queue
			xray.Capture(ctx, "SendMsg", func(ctx1 context.Context) error {
				_, err := svc.SendMessageWithContext(ctx, &sqs.SendMessageInput{
					MessageBody: aws.String(ri),
					QueueUrl:    aws.String(q)})

				// print error if it occurs
				if err != nil {
					log.Println(err)
				}
				xray.AddMetadata(ctx, "SentMessages", "10")

				return nil
			})

			// increase total count
			tot++

			// print log line
			log.Println(strconv.Itoa(tot) + " " + ri)

			// close xray segments
			Seg.Close(nil)
		}
	}

	// print total messages completed
	fmt.Println("sent " + strconv.Itoa(tot) + " messages")
}

// start lambda handler
func main() {
	lambda.Start(handler)
}
