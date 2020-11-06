package dnsbl

import (
	//"encoding/json"
	"github.com/egreen64/codingchallenge/config"
	"github.com/nerdbaggy/godnsbl"
)

//DnsblReturn type
type DnsblReturn godnsbl.DnsblReturn

//DnsblData type
type DnsblData godnsbl.DnsblData

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
func (d *Dnsbl) Lookup(ipAddress string) DnsblReturn {
	godnsbl.BlacklistDomains = d.BlocklistDomains
	resp := godnsbl.CheckBlacklist(ipAddress)

	return DnsblReturn(resp)
}
