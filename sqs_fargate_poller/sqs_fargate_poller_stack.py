from aws_cdk import core
from aws_cdk import aws_ec2 as ec2
from aws_cdk import aws_ecs as ecs
from aws_cdk import aws_sqs as sqs
from aws_cdk import aws_ecs_patterns as ecs_patterns
from aws_cdk import aws_lambda as lambda_

class SQSStack(core.Stack):

    def __init__(self, scope: core.Construct, id: str, **kwargs) -> None:
        super().__init__(scope, id, **kwargs)

        # build the docker image from local './docker/' directory
        container = ecs.ContainerImage.from_asset(directory = 'docker')

        # create a new VPC
        vpc = ec2.Vpc(
            self, "Vpc",
            max_azs = 10
        )

        # create a new ECS cluster
        cluster = ecs.Cluster(self, "FargateSQS", vpc = vpc)

        # create a new SQS queue
        msg_queue = sqs.Queue(self, 'SQSQueue',
                                         visibility_timeout=core.Duration.seconds(300),
                                         queue_name='SQSQueue')

        # create the queue processing service on fargate with a locally built container
        # the pattern automatically adds an environment variable with the queue name for the container to read
        queue_processing_fargate_service = ecs_patterns.QueueProcessingFargateService(self, "Service",
            cluster = cluster,
            memory_limit_mib = 512,
            cpu = 256,
            image = container,
            enable_logging = True,
            desired_task_count = 1,
            max_scaling_capacity = 3,
            queue = msg_queue
        )

        # create a lambda function to generate load
        sqs_lambda = lambda_.Function(self, "GenerateLoadSQS",
            runtime = lambda_.Runtime.PYTHON_3_8,
            code = lambda_.Code.asset("lambda"),
            handler = "lambda.handler",
            timeout = core.Duration.seconds(60),
            environment = {'SQS_QUEUE': msg_queue.queue_url}
        )
        
        # add the Lambda IAM permission to send SQS messages
        msg_queue.grant_send_messages(sqs_lambda)
