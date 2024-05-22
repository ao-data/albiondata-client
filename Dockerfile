#FROM ubuntu:latest
FROM golang:1.16.6-buster



RUN apt-get update
RUN apt-get install -y make sudo nsis ca-certificates openssl

RUN mkdir -p /go/src/github.com/ao-data/albiondata-client

RUN mkdir -p /go/src/github.com/ao-data/albiondata-client
WORKDIR /go/src/github.com/ao-data/albiondata-client
COPY . .
