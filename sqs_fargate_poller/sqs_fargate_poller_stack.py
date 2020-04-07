import os, shutil, sys, subprocess

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

from aws_cdk.aws_iam import PolicyStatement

class SQSStack(core.Stack):

    def __init__(self, scope: core.Construct, id: str, **kwargs) -> None:
        super().__init__(scope, id, **kwargs)

        # build the docker image from local "./docker/" directory
        sqscontainer = aws_ecs.ContainerImage.from_asset(directory = "docker")

        # add the aws-xray-daemon as a sidecar running on UDP/2000
        xraycontainer = aws_ecs.ContainerImage.from_registry("amazon/aws-xray-daemon")

        # create a new VPC with the max amount of AZ"s and one NAT gateway to reduce static infra cost
        # TODO - include VPC SQS endpoint so the NAT gateway isn"t needed anymore
        vpc = aws_ec2.Vpc(
            self, "Vpc",
            max_azs = 10,
            nat_gateways = 3,
            subnet_configuration = [
                aws_ec2.SubnetConfiguration(
                    name = "private", cidr_mask = 24, subnet_type = aws_ec2.SubnetType.PRIVATE
                ),
                aws_ec2.SubnetConfiguration(
                    name = "public", cidr_mask = 28, subnet_type = aws_ec2.SubnetType.PUBLIC
                )
            ]
        )

        # create a new ECS cluster
        cluster = aws_ecs.Cluster(self, "FargateSQS", vpc = vpc)

        # create a new SQS queue
        msg_queue = aws_sqs.Queue(self, "SQSQueue",
            visibility_timeout = core.Duration.seconds(0),
            retention_period = core.Duration.minutes(30)
        )

        # create the queue processing service on fargate with a locally built container
        # the pattern automatically adds an environment variable with the queue name for the container to read
        fargate_service = aws_ecs_patterns.QueueProcessingFargateService(self, "Service",
            cluster = cluster,
            memory_limit_mib = 512,
            cpu = 256,
            image = sqscontainer,
            enable_logging = True,
            desired_task_count = 0,
            max_scaling_capacity = 5,
            scaling_steps = [
                {"upper": 0, "change": -5}, 
                {"lower": 1, "change": +1}, 
                {"lower": 50000, "change": +2}, 
                {"lower": 250000, "change": +4}
            ],
            queue = msg_queue,
            environment = {
                "sqs_queue_url": msg_queue.queue_url
            }
        )

        # add the standard aws xray sidecar to the container task
        xray_sidecar = fargate_service.task_definition.add_container(
            "xraycontainer",
            image = xraycontainer,
            logging = fargate_service.log_driver
        )

        # expose the sidecar on port UDP/2000
        xray_sidecar.add_port_mappings(aws_ecs.PortMapping(container_port = 2000, protocol = aws_ecs.Protocol.UDP))

        # build the go binary for the lambda SQS generator
        # since CDK cannot natively build Go binaries yet, we need to do this manually
        src_file    = './loadgen/loadgensqs.go'
        src_md5     = subprocess.check_output(['md5', '-q', src_file])

        build_file    = './loadgen/tempbuilddir/tmp.go'
        build_md5     = subprocess.check_output(['md5', '-q', build_file])

        # check if the md5's of go source code are different
        if src_md5 != build_md5:

            print("\ndifferent files detected : \n" + src_file + " " + str(src_md5) + "\n" + build_file + " " + str(build_md5)+"\n")
            print('\nbuilding go binary from ' + src_file +' as it changed since last cdk deploy...\n')
            
            # copy the source file from ./loadgen to the local directory
            os.system("cp " + src_file + " .")

            # build the go binary in root and copy it to the ./bin directory
            os.system('go get -d -v .')
            build = os.system('GOARCH=amd64 GOOS=linux go build -o loadgensqs loadgensqs.go')
            
            print('build '+str(build), type(build))

            # if build didn't exit cleanly, there likely was an issue with the go build and a break is returned
            if build == 0:

                # if upx is installed, compress the binary to reduce size
                if shutil.which('upx'):
                    print('\ncompressing go binary with upx...\n')
                    os.system('upx -9 loadgensqs && chmod +x loadgensqs')

                # create a zip file for the lambda go deployment
                print('\ncreate zip file for lambda deployment and move to ./loadgen dir\n')
                os.system('zip -r9 lambda.zip loadgensqs')
                os.system('mv lambda.zip ./loadgen/')

                # move the loadgensqs binary and go source code to ./tempbuilddir for caching
                os.system('mv loadgensqs ./loadgen/tempbuilddir/')
                os.system('mv loadgensqs.go ./loadgen/tempbuilddir/')
            
            else:
                print('\nfound error in go source code, please see build error above\n')

                # remove go source code from root dir
                os.system('rm ./loadgensqs.go')
                raise SystemExit
            
        # create a lambda function to generate load
        sqs_lambda = aws_lambda.Function(self, "GenerateLoadSQS",
            runtime = aws_lambda.Runtime.GO_1_X,
            code = aws_lambda.Code.asset("./loadgen/lambda.zip"),
            handler = "loadgensqs",
            timeout = core.Duration.seconds(10),
            memory_size = 256,
			retry_attempts = 0,
            tracing = aws_lambda.Tracing.ACTIVE,
            environment = {
                "sqs_queue_url": msg_queue.queue_url,
                "total_message_count": "100"
            }
        )
        
        # create a new cloudwatch rule running every hour to trigger the lambda function
        eventRuleHour   = aws_events.Rule(self, "lambda-generator-hourly-rule",
            enabled     = True,
            schedule    = aws_events.Schedule.cron(minute = "0"))

        eventRuleHour.add_target(aws_events_targets.LambdaFunction(sqs_lambda))

        # create a new cloudwatch rule running every minute to trigger the lambda function
        eventRuleMinu   = aws_events.Rule(self, "lambda-generator-minute-rule",
            enabled     = True,
            schedule    = aws_events.Schedule.cron(minute = "*"))

        eventRuleMinu.add_target(aws_events_targets.LambdaFunction(sqs_lambda))

        # add the Lambda IAM permission to send SQS messages
        msg_queue.grant_send_messages(sqs_lambda)

        # add XRay permissions to Fargate task
        xray_policy = PolicyStatement(
            resources = ["*"],
            actions = ["xray:GetGroup",
                     "xray:GetGroups",
                     "xray:GetSampling*",
                     "xray:GetTime*",
                     "xray:GetService*",
                     "xray:PutTelemetryRecords",
                     "xray:PutTraceSegments"]
        )

        fargate_service.task_definition.add_to_task_role_policy(xray_policy)
        #sqs_lambda.add_to_role_policy(xray_policy)
