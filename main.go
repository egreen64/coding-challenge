package main

import (
	"encoding/json"
	"log"

	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
)

func main() {
	//Read Config File
	config := config.GetConfig()

	db.NewDatabase(config)

	//Instantiate DNS Blacklist instance
	dnsbl := dnsbl.NewDnsbl(config)

	resp := dnsbl.Lookup("127.0.0.2")
	jsonResponse, _ := json.Marshal(resp)
	log.Printf("%s\n", jsonResponse)
}
