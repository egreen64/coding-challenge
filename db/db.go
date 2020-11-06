package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/graph/model"
	"github.com/egreen64/codingchallenge/utils"

	_ "github.com/mattn/go-sqlite3" //Required for sql driver registration
)

//Database type
type Database struct {
	dbPath string
	db     *sql.DB
}

//NewDatabase instantiate database instance
func NewDatabase(config *config.File) *Database {
	log.Printf("opening database type: %s, database name: %s\n", config.Database.DbType, config.Database.DbPath)

	if !config.Database.Persist {
		os.Remove(config.Database.DbPath)
	}

	db, err := sql.Open(config.Database.DbType, config.Database.DbPath)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatalf("unable to open database data source %s, error:%s\n", config.Database.DbPath, err)
	}
	if !utils.FileExists(config.Database.DbPath) {
		sqlStmt := `
			CREATE TABLE IF NOT EXISTS dns_blocklist (
				ip_address TEXT PRIMARY KEY NOT NULL, 
				id TEXT NOT NULL, 
				response_code text,
				created_at DATETIME CURRENT_TIMESTAMP, 
				updated_at DATETIME CURRENT_TIMESTAMP
			);
		`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			log.Printf("unable to create database table dns_blocklist, error: %s\n", err)
			log.Fatalf("create table sql statement: %s\n", sqlStmt)
		}
	}

	log.Printf("datbase %s succesfully opened\n", config.Database.DbPath)

	dbi := Database{
		dbPath: config.Database.DbPath,
		db:     db,
	}

	return &dbi
}

//CloseDatabase function
func (db *Database) CloseDatabase() {
	log.Printf("closing database %s\n", db.dbPath)
	db.db.Close()
	log.Printf("database %s closed\n", db.dbPath)
}

//UpsertRecord function
func (db *Database) UpsertRecord(record *model.DNSBlockListRecord) error {

	sqlStmt := `
		INSERT INTO dns_blocklist(
			id,
			ip_address,
			response_code,
			created_at,
			updated_at
		) values(?, ?, ?, ?, ?)
		ON CONFLICT(ip_address) DO UPDATE SET
			response_code = ?,
			updated_at = ?
	`

	stmt, err := db.db.Prepare(sqlStmt)
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	currentTime := time.Now().Format(time.RFC3339)
	_, err = stmt.Exec(record.UUID, record.IPAddress, record.ResponseCode, currentTime, currentTime, record.ResponseCode, currentTime)
	if err != nil {
		err = fmt.Errorf("unexpected database insert error for ip address %s, error: %s", record.IPAddress, err)
		log.Printf("%s\n", err)
	}

	return err
}

//SelectRecord function
func (db *Database) SelectRecord(ipAddress string) (*model.DNSBlockListRecord, error) {
	sqlStmt := `
		SELECT
			id,
			ip_address,
			response_code,
			created_at,
			updated_at
		FROM dns_blocklist 
		WHERE ip_address = ?
	`

	stmt, err := db.db.Prepare(sqlStmt)
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(ipAddress)
	if err != nil {
		log.Printf("unexpected database select exec error for ip address %s, error: %s", ipAddress, err)
		err := fmt.Errorf("unexpected query failure encountered for ip address %s", ipAddress)
		return nil, err
	}
	var dblRec model.DNSBlockListRecord
	var createdAt string
	var updatedAt string

	err = db.db.QueryRow(sqlStmt, ipAddress).Scan(
		&dblRec.UUID,
		&dblRec.IPAddress,
		&dblRec.ResponseCode,
		&createdAt,
		&updatedAt,
	)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("no record found in database select for ip address %s, error: %s", ipAddress, err)
		err := fmt.Errorf("blocklist for ip address %s not found", ipAddress)

		return nil, err
	case err != nil:
		log.Printf("unexpected database select error for ip address %s, error: %s", ipAddress, err)
		err := fmt.Errorf("unexpected query failure encountered for ip address %s", ipAddress)
		log.Fatalf("query error: %v\n", err)
	default:
		dblRec.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		dblRec.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	}

	return &dblRec, nil
}
