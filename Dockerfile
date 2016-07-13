FROM golang:latest
MAINTAINER xtaci <daniel820313@gmail.com>
ENV GOBIN /go/bin
COPY src /go/src
WORKDIR /go
RUN go install agent
RUN rm -rf pkg src
ENTRYPOINT /go/bin/agent
EXPOSE 8888 8888/udp
