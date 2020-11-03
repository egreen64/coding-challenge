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
	blackListDomains []string
}

//NewDnsbl - Create DNS BlackList instance
func NewDnsbl(config *config.File) *Dnsbl {
	dnsbl := Dnsbl{
		blackListDomains: config.Dnsbl.BlacklistDomains,
	}
	return &dnsbl
}

//Lookup - Blacklist Domain Lookup
func (d *Dnsbl) Lookup(ipAddress string) DnsblReturn {
	godnsbl.BlacklistDomains = d.blackListDomains
	resp := godnsbl.CheckBlacklist(ipAddress)
	return DnsblReturn(resp)
}
