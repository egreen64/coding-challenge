#!/bin/bash 

go run github.com/99designs/gqlgen generate
go build
go test -v ./...
rm -rf coding_challenge.db
rm -rf coding_challenge.db-journal
rm -rf coding_challenge.log
