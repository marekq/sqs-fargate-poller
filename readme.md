sqs-fargate-poller
==================

Deploy an SQS queue triggered Fargate container using the AWS CDK. In addition, a load generator Lambda is included which puts random messages on the SQS queue. 


![alt text](./docs/diagram.svg)


Installation
------------

Run 'cdk deploy' in the main directory. The Docker container and Lambda function will be built and deployed based on the locally stored sourcecode. 


Roadmap
-------

- [ ] Rewrite SQS generation Lambda to Golang (once CDK properly supports this or using a workaround to run 'go build')
- [ ] Change alarm metric monitoring rate from every 5 minutes to every minute, so that Fargate scales more accurately depending on the amount of messages on the queue. 


Contact
-------

In case you have any suggestions, questions or remarks, please raise an issue or reach out to @marekq.

