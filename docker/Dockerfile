FROM amazonlinux:latest
RUN yum install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm && yum update -y
RUN yum install python3-pip -y && pip3 install --user boto3

ADD dockerqueue.py /tmp
CMD ["python3", "-u", "/tmp/dockerqueue.py"]
