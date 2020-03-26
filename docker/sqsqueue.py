import boto3
import os

sqs = boto3.client('sqs')

# get the SQS queue name
qenv = os.environ['QUEUE_NAME']
print('reading messages from '+qenv)

# get the SQS queue URL
qurl = sqs.get_queue_url(QueueName = qenv)['QueueUrl']

# read the messages from queue and wait 20 seconds
def readmsg():
    response = sqs.receive_message(
        QueueUrl = qurl,
        MessageAttributeNames = ['All'],
        MaxNumberOfMessages = 1,
        WaitTimeSeconds = 20
    )

    if 'Messages' in response:        
        # get the message receipt handle
        receipth = response['Messages'][0]['ReceiptHandle']

        # delete the message from the queue
        delete = sqs.delete_message(
            QueueUrl = qurl, 
            ReceiptHandle = receipth)

        print(str(response)+'\n')
    else:
        print('no match '+str(response)+'\n')

###

print('starting read from '+qurl)

# keep running readmsg() indefinetly
while True:
    readmsg()