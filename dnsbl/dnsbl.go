package dnsbl

import (
	//"encoding/json"
	"github.com/egreen64/codingchallenge/config"
	"github.com/nerdbaggy/godnsbl"
)

//Return type
type Return godnsbl.DnsblReturn

//Data type
type Data godnsbl.DnsblData

//Dnsbl instance type
type Dnsbl struct {
	BlocklistDomains []string
}

//NewDnsbl - Create DNS Blocklist instance
func NewDnsbl(config *config.File) *Dnsbl {
	dnsbl := Dnsbl{
		BlocklistDomains: config.Dnsbl.BlocklistDomains,
	}
	return &dnsbl
}

//Lookup - Blocklist Domain Lookup
func (d *Dnsbl) Lookup(ipAddress string) Return {
	godnsbl.BlacklistDomains = d.BlocklistDomains
	resp := godnsbl.CheckBlacklist(ipAddress)

	return Return(resp)
}
