package graph

import (
	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

//Resolver Type
type Resolver struct {
	Config   *config.File
	Database *db.Database
	DNSBL    *dnsbl.Dnsbl
}
