package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// set the poll timeout value in seconds and the SQS queue URL
var (
	polltimeout = aws.Int64(20)
	sqsUrl      = os.Getenv("sqs_queue_url")
)

// retrieve a message from SQS
func pollmsg(svc *sqs.SQS) {
	result, _ := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(sqsUrl),
		MaxNumberOfMessages:   aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{"All"}),
		WaitTimeSeconds:       polltimeout,
	})

	// print the message and delete it from queue
	if result.Messages != nil {
		msghandler := result.Messages[0].ReceiptHandle
		fmt.Println(result.Messages)

		_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(sqsUrl),
			ReceiptHandle: msghandler,
		})

		// print error if the message can't be deleted
		if err != nil {
			fmt.Println("message not deleted")
		}
	}
}

// connect to SQS and run an infinite loop on pollers
func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	fmt.Println("sqsUrl " + sqsUrl)
	svc := sqs.New(sess)

	for {
		pollmsg(svc)
	}
}
