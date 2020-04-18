# build the go binary in a golang build container
# download andd install the time and upx packages 
FROM golang as builder
RUN apt-get update -y && apt-get install time upx zip -y --no-install-recommends

# copy the go file to /tmp
WORKDIR /tmp
ADD loadgensqs.go .

# retrieve go dependancies and build the binary
RUN go get -d -v .
RUN CGO_ENABLED=0 go build -o loadgensqs

# compress the go binary to reduce its size by ~40% and ~5MB in size
# this adds less than 10 seconds to the build time but speeds up the lambda cold start time
RUN time upx -9 loadgensqs

# create a zip with the lambda go binary and lambda source code for backup
RUN zip -r9 lambda.zip loadgensqs loadgensqs.go
