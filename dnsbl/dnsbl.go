package dnsbl

import (
	//"encoding/json"
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
func NewDnsbl(domains []string) *Dnsbl {
	dnsbl := Dnsbl{
		blackListDomains: domains,
	}
	return &dnsbl
}

//Lookup - Blacklist Domain Lookup
func (d *Dnsbl) Lookup(ipAddress string) DnsblReturn {
	godnsbl.BlacklistDomains = d.blackListDomains
	resp := godnsbl.CheckBlacklist(ipAddress)
	return DnsblReturn(resp)
}
