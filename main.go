package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/google/uuid"
)

func main() {
	//Read Config File
	config := config.GetConfig()

	dbi := db.NewDatabase(config)

	//Instantiate DNS Blacklist instance
	dnsbl := dnsbl.NewDnsbl(config)

	ipAddress := "127.0.0.2"
	resp := dnsbl.Lookup(ipAddress)
	jsonResponse, _ := json.Marshal(resp)
	log.Printf("%s\n", jsonResponse)

	dnsBlacklistRecord := db.DNSBlacklistRecord{
		ID:           uuid.New().String(),
		IPAddress:    ipAddress,
		ResponseCode: resp.Responses[0].Resp,
	}

	dbi.UpsertRecord(&dnsBlacklistRecord)
	dblRec, _ := dbi.SelectRecord(ipAddress)
	log.Println(dblRec)

	time.Sleep(2 * time.Second)

	dnsBlacklistRecord.ResponseCode = "NXDOMAIN"
	dbi.UpsertRecord(&dnsBlacklistRecord)
	dblRec, _ = dbi.SelectRecord(ipAddress)
	log.Println(dblRec)

	dbi.CloseDatabase()
}
