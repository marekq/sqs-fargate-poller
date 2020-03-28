import boto3, botocore, os, random, queue, threading

# configure total message count to send and  the amount of threads
msgc	= int(os.environ['total_message_count'])
thrc	= int(os.environ['python_worker_threads'])

# retrieve SQS queue URL and set up connection to SQS service
qurl    = os.environ['sqs_queue_url']
sqs     = boto3.client('sqs', config = botocore.client.Config(max_pool_connections = thrc))

# set counter values
count 	= 0
fail 	= 0

# create a queue
q1     	= queue.Queue()


# worker for queue jobs
def worker():
	while not q1.empty():
		send(q1.get())
		q1.task_done()


# send a message, increase the success or failure counter
def send(x):
	try:
		sqs.send_message(QueueUrl = qurl, MessageBody = x)
		global count
		count += 1

	except Exception as e:
		global fail
		fail += 1
		print('error in send(): '+str(e))


# lambda handler
def handler(event, context):
	print('sending '+str(msgc)+' messages to '+qurl)

	# generate randomint messages and put them on the queue
	for x in range(int(msgc)):
		q1.put(str(random.randint(1000, 9999)))

	# start the processing threads
	for x in range(int(thrc)):
		t = threading.Thread(target = worker)
		t.daemon = True
		t.start()
	q1.join()

	# print how many messages were sent and exit
	print('sucessfully sent '+str(count)+' messages')
	print('failed to send '+str(fail)+' messages')
