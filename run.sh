#!/bin/bash

if [[ $# -eq 0 ]] ; then
    port=8080
else
    port=$1
fi

echo docker run -d -p $port:8080 --name coding_challenge coding_challenge:latest
