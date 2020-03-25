#!/usr/bin/env python3

from aws_cdk import core

from sqs_fargate_poller.sqs_fargate_poller_stack import SQSStack

app = core.App()
SQSStack(app, "sqs-fargate-processor")

app.synth()
