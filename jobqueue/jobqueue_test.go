package jobqueue

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/stretchr/testify/require"
)

func TestJobqueue(t *testing.T) {
	//Disable logging
	log.SetOutput(ioutil.Discard)

	//Set location of config file
	os.Setenv("GO_CONFIG", "../config.json")

	//Get config file
	config := config.GetConfig()

	//Initialize databse
	database := db.NewDatabase(config)

	//Instantiate DNS Blocklist instance
	dnsbl := dnsbl.NewDnsbl(config)

	//Instantiage job queue
	jobQueue := NewJobQueue(config, dnsbl, database)

	t.Run("new_jobqueue_success", func(t *testing.T) {

		newJobQueue := NewJobQueue(config, dnsbl, database)
		require.NotEqual(t, nil, newJobQueue)
	})

	t.Run("stop_jobqueue_success", func(t *testing.T) {

		newJobQueue := NewJobQueue(config, dnsbl, database)
		require.NotEqual(t, nil, newJobQueue)

		resp := newJobQueue.Stop()

		require.Equal(t, true, resp)
	})

	t.Run("add_jobqueue_success", func(t *testing.T) {

		resp := jobQueue.AddJob([]string{"127.0.0.1"})
		require.Equal(t, true, resp)
	})

	t.Run("add_jobqueue_failure_queue_full", func(t *testing.T) {

		var resp bool
		for {
			resp = jobQueue.AddJob([]string{"127.0.0.1"})
			if !resp {
				break
			}
		}
		require.Equal(t, false, resp)
	})
}
