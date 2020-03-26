from aws_cdk import (
    core,
    aws_ec2,
    aws_ecs,
    aws_lambda,
    aws_sqs,
    aws_ecs_patterns,
    aws_events,
    aws_events_targets
)

class SQSStack(core.Stack):

    def __init__(self, scope: core.Construct, id: str, **kwargs) -> None:
        super().__init__(scope, id, **kwargs)

        # build the docker image from local './docker/' directory
        container = aws_ecs.ContainerImage.from_asset(directory = 'docker')

        # create a new VPC
        vpc = aws_ec2.Vpc(
            self, "Vpc",
            max_azs = 10
        )

        # create a new ECS cluster
        cluster = aws_ecs.Cluster(self, "FargateSQS", vpc = vpc)

        # create a new SQS queue
        msg_queue = aws_sqs.Queue(self, 'SQSQueue',
            visibility_timeout = core.Duration.seconds(0),
            retention_period = core.Duration.minutes(10)
        )

        # create the queue processing service on fargate with a locally built container
        # the pattern automatically adds an environment variable with the queue name for the container to read
        queue_processing_fargate_service = aws_ecs_patterns.QueueProcessingFargateService(self, "Service",
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
        sqs_lambda = aws_lambda.Function(self, "GenerateLoadSQS",
            runtime = aws_lambda.Runtime.PYTHON_3_8,
            code = aws_lambda.Code.asset("lambda"),
            handler = "lambda.handler",
            timeout = core.Duration.seconds(60),
            memory_size = 512,
            environment = {'SQS_QUEUE': msg_queue.queue_url},
            tracing = aws_lambda.Tracing.ACTIVE
        )
        
        # create a new cloudwatch rule running every minute and triggers the lambda function
        eventRule = aws_events.Rule(self, 'lambda-generator-1min-rule',
            enabled = True,
            schedule = aws_events.Schedule.cron( minute = '*' ))
        eventRule.add_target(aws_events_targets.LambdaFunction(sqs_lambda))

        # add the Lambda IAM permission to send SQS messages
        msg_queue.grant_send_messages(sqs_lambda)
