package db

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/graph/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDb(t *testing.T) {
	//Disable logging
	log.SetOutput(ioutil.Discard)

	//Set location of config file
	os.Setenv("GO_CONFIG", "../config.json")

	//Get config file
	config := config.GetConfig()

	//Delete database
	os.Remove(config.Database.DbPath)

	var db *Database

	t.Run("new_database_success", func(t *testing.T) {

		db = NewDatabase(config)
		require.NotEqual(t, nil, db)

		os.Remove(config.Database.DbPath)
	})

	t.Run("close_database_success", func(t *testing.T) {

		db = NewDatabase(config)
		require.NotEqual(t, nil, db)

		db.CloseDatabase()
		os.Remove(config.Database.DbPath)
	})
	t.Run("upsert_record_success", func(t *testing.T) {

		db = NewDatabase(config)
		require.NotEqual(t, nil, db)

		record := model.DNSBlockListRecord{
			UUID:         uuid.New().String(),
			IPAddress:    "127.0.0.12",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			ResponseCode: "NXDOMAIN",
		}
		err := db.UpsertRecord(&record)
		require.Equal(t, nil, err)

		db.CloseDatabase()
		os.Remove(config.Database.DbPath)
	})
}
