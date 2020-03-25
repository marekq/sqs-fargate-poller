import boto3, os, random

qurl = os.environ['SQS_QUEUE']

sqs = boto3.client('sqs')

def handler(event, context):
    for x in range(100):
        print('sending '+str(x))
        sqs.send_message(QueueUrl = qurl, MessageBody = str(random.randint(1000, 9999)))

