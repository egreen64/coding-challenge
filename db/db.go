package db

import (
	"database/sql"
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
	log.Printf("database Type: %s, Database Name: %s\n", config.Database.DbType, config.Database.DbPath)

	if !config.Database.Persist {
		os.Remove(config.Database.DbPath)
	}

	db, err := sql.Open(config.Database.DbType, config.Database.DbPath)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal("unable to use data source name", err)
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
			log.Printf("%q: %s\n", err, sqlStmt)
		}
	}

	dbi := Database{
		dbPath: config.Database.DbPath,
		db:     db,
	}

	return &dbi
}

//CloseDatabase function
func (db *Database) CloseDatabase() {
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
		log.Printf("insert error: %s\n", err)
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
		log.Printf("select error: %s\n", err)
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
		log.Printf("no row with id %s\n", ipAddress)
		return nil, err
	case err != nil:
		log.Fatalf("query error: %v\n", err)
	default:
		dblRec.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		dblRec.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	}

	return &dblRec, nil
}

//UpdateRecord function
func (db *Database) UpdateRecord(record *model.DNSBlockListRecord) error {
	sqlStmt := `
	UPDATE dns_blocklist set
		response_code = ?,
		updated_at = ?
		where ip_address = ?
	`

	stmt, err := db.db.Prepare(sqlStmt)
	if err != nil {
		panic(err)
	}

	defer stmt.Close()

	currentTime := time.Now().Format(time.RFC3339)
	_, err = stmt.Exec(record.ResponseCode, currentTime, record.IPAddress)
	if err != nil {
		log.Printf("update error: %s\n", err)
	}

	return err
}
