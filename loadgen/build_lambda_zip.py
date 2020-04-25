import os, subprocess

# calculate the md5 values of the source code and the temporary build file
src_file    = './loadgen/loadgensqs.go'
src_md5     = subprocess.check_output(['md5', '-q', src_file])

build_file  = './loadgen/tempbuilddir/tmp.go'
build_md5   = subprocess.check_output(['md5', '-q', build_file])

# check if the md5's of go source code are different
if src_md5 != build_md5:
    print("\nLAMBDA - different file content detected between:")
    print("\ncurrent lambda sourcecode - " + str(src_md5) + " " + src_file)
    print("\ncached build sourcecode   - " + str(build_md5)+ " " + build_file)
    print("\nbuilding go binary in Docker container from " + src_file + "\n")
    
    # run the docker image build using the local Dockerfile
    os.system("cd loadgen/ && DOCKER_BUILDKIT=1 docker build -f Dockerfile -t generateloadsqslambda . ")
    os.system("docker create --name generateloadsqslambda generateloadsqslambda:latest")

    # get the docker image id of the container
    docker_id = subprocess.check_output("docker ps -a | grep generateloadsqslambda | head -1 | awk {'print $1'}", shell = True).strip().decode('utf-8')
    print("\nbuilt and created container with id " + str(docker_id))

    # copy the go binary from the container to the ./loadgen directory
    os.system("docker cp generateloadsqslambda:/tmp/lambda.zip ./loadgen/lambda.zip")
    os.system("cp "+ src_file + " " + build_file)
    print("\ncopied lambda.zip from container to ./loadgen/lambda.zip")

    # delete the docker container from the host
    print("\ndeleting container with id " + str(docker_id))
    os.system("docker rm " + docker_id)

else:
    print("\nLAMBDA - no different file content found in " + src_file + ", exiting")
