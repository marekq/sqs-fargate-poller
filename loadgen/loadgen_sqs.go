package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	// region
	r = "eu-west-1"

	// amount of puts per go routine
	c = 1000

	// total amount of go routines
	d = 10
)

//
// do not change anything below this line
//

func main() {
	q := os.Getenv("sqs_queue_url")
	fmt.Println("sending to " + q)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// create a session with sqs
	svc1 := sqs.New(sess)

	// keep track of total sent messages
	tot := 0

	for a := 0; a < c; a++ {

		// create a wait group to wait for go subroutines
		var wg sync.WaitGroup

		// spawn go routines depending on total count
		for b := 0; b < d; b++ {

			// add one count to the workgroup
			wg.Add(1)

			// run the send message command in parallel
			go func() {
				defer wg.Done()
				ri := strconv.Itoa(rand.Intn(9999))

				// send the message to the sqs queue
				_, err := svc1.SendMessage(&sqs.SendMessageInput{MessageBody: aws.String(ri), QueueUrl: aws.String(q)})

				if err != nil {
					log.Println(err)
				}

				tot += 1
				log.Println(strconv.Itoa(tot) + " " + ri)

			}()

		}
		// wait for all routines to finish
		wg.Wait()

	}
}
