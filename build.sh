#!/bin/bash 

rev=`git rev-parse HEAD | cut -c1-8`
echo ". cross-compiling codingchallange commit version $rev"

rm -rf ./coding_challenge.log
rm -rf ./coding_challenge.db
rm -rf ./codingchallenge

#build and test coding challenge project
docker run --rm -v "$PWD":/usr/src/codingchallenge -w /usr/src/codingchallenge -e GOOS=linux golang:1.15.3 go build;go test -v ./...

#build docker image
docker build -t coding_challenge:latest .

#push docker image
#docker tag coding_challenge:latest egreen6464/coding_challenge
#docker push egreen6464/coding_challenge:latest
