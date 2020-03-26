import boto3, botocore, os, random, queue, threading

qurl    = os.environ['SQS_QUEUE']
sqs     = boto3.client('sqs', config = botocore.client.Config(max_pool_connections = 20))

# create a queue
q1     	= queue.Queue()

# worker for queue jobs
def worker():

	while not q1.empty():
		send(q1.get())
		q1.task_done()

	#print('completed thread with 100 msg')


# send a message
def send(x):
	sqs.send_message(QueueUrl = qurl, MessageBody = x)
 

# lambda handler
def handler(event, context):
	for x in range(5000):
		q1.put(str(random.randint(1000, 9999)))

	# start 20 threads
	for x in range(50):
		t = threading.Thread(target = worker)
		t.daemon = True
		t.start()
	q1.join()
