# build the go binary in a golang build container
# download andd install the time and upx packages 
FROM golang AS builder
RUN apt-get update -y && apt-get install upx time -y --no-install-recommends

# copy the go file to /tmp
WORKDIR /tmp
ADD dockerqueue.go /tmp

# retrieve go dependancies and build the binary
RUN go get -d -v .
RUN CGO_ENABLED=0 go build -o dockerqueue

# compress the go binary to reduce its size by ~40% and ~5MB in size
# this adds less than 10 seconds to the build time but speed up the container start time
RUN time upx -9 dockerqueue

# build the actual container with alpine, which will be scanned by ECR for vulnerabilities
FROM alpine
COPY --from=builder /tmp/dockerqueue /tmp/dockerqueue

# start the sqs queue processor once the container starts and run indefinitely
CMD ["/tmp/dockerqueue"]
