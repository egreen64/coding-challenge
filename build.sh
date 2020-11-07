#!/bin/bash 

rev=`git rev-parse HEAD | cut -c1-8`
echo ". cross-compiling codingchallange commit version $rev"

rm -rf ./coding_challenge.log
rm -rf ./coding_challenge.db
rm -rf ./codingchallenge

docker run --rm -v "$PWD":/usr/src/codingchallenge -w /usr/src/codingchallenge -e GOOS=linux golang:1.15.3 go build

docker build -t coding_challenge:latest .
