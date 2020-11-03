package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/egreen64/codingchallenge/config"
	_ "github.com/mattn/go-sqlite3"
)

//Database type
type Database struct {
	db *sql.DB
}

//NewDatabase instantiate database instance
func NewDatabase(config *config.File) *Database {
	log.Printf("Database Type: %s, Database Name: %s\n", config.Database.DbType, config.Database.DbName)
	db, err := sql.Open(config.Database.DbType, config.Database.DbName)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal("unable to use data source name", err)
	}
	if !dbExists(config.Database.DbName) {
		sqlStmt := `
		create table foo (id integer not null primary key, name text);
		delete from foo;
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("%q: %s\n", err, sqlStmt)
		}
	}

	dbi := Database{
		db: db,
	}
	defer db.Close()
	return &dbi
}

func dbExists(dbName string) bool {
	info, err := os.Stat(dbName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
