import aws_xray_sdk, boto3, botocore, json, logging, os, random, queue, threading
from aws_xray_sdk.core import patch_all, xray_recorder

# patch libraries for xray tracing
patch_all()

# configure total message count to send and  the amount of threads
msgc	= int(os.environ['total_message_count'])
thrc	= int(os.environ['python_worker_threads'])

# retrieve SQS queue URL and set up connection to SQS service
qurl    = os.environ['sqs_queue_url']
sqs     = boto3.client('sqs', config = botocore.client.Config(max_pool_connections = thrc))

# set counter values
count 	= 0
fail 	= 0
total 	= 0

# create a queue
q1     	= queue.Queue()


# worker for queue jobs
def worker():
	while not q1.empty():
		send(q1.get())
		q1.task_done()

		global total
		total += 10

		# print sent count per 10000 messages for debugging purposes
		if (total % 10000 == 0):
			print("sent "+str(total)+" messages")


# send a message, increase the success or failure counter
@xray_recorder.capture("send_message")
def send(x):
	msgs 	= []

	try:
		# generate 10 random messages 
		for z in range(10):
			msg = {
				'Id': str(z),
				'MessageBody': json.dumps(x),
				'DelaySeconds': 0
			}
			msgs.append(msg)

		# send the batch of messages to SQS
		sqs.send_message_batch(QueueUrl = qurl, Entries = msgs)
		global count
		count += 10

	# print the error in case the message cannot be sent succesfully
	except Exception as e:
		global fail
		fail += 10
		print('error in send(): '+str(e))


# lambda handler
def handler(event, context):
	logging.basicConfig(level = 'DEBUG')
	logging.getLogger('aws_xray_sdk').setLevel(logging.WARNING)
	xray_recorder.configure(context_missing = "RUNTIME_ERROR")

	xray_recorder.begin_subsegment('start_lambda')
	print('sending '+str(msgc)+' messages to '+qurl)

	# generate randomint messages and put them on the queue
	for x in range(int(msgc / 10)):
		q1.put(str(random.randint(1000, 9999)))

	xray_recorder.end_subsegment()

	# start the processing threads
	xray_recorder.begin_subsegment('running_threads')

	for x in range(int(thrc)):
		t = threading.Thread(target = worker)
		t.daemon = True
		t.start()
	q1.join()

	xray_recorder.end_subsegment()

	# print how many messages were sent and exit
	xray_recorder.begin_subsegment('print_results')

	print('sucessfully sent '+str(count)+' messages')
	print('failed to send '+str(fail)+' messages')
	xray_recorder.end_subsegment()
