package config

//File struct
type File struct {
	Database Database `json:"db"`
	Dnsbl    Dnsbl    `json:"dnsbl"`
}

//Database type
type Database struct {
	DbType string `json:"db_type"`
	DbName string `json:"db_name"`
}

//Dnsbl type
type Dnsbl struct {
	BlacklistDomains []string `json:"blacklist_domains"`
}
