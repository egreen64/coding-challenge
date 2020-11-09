package jobqueue

import (
	"log"
	"sync"
	"time"

	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/egreen64/codingchallenge/graph/model"
	"github.com/google/uuid"
)

//JobQueue type
type JobQueue struct {
	dnsbl       *dnsbl.Dnsbl
	db          *db.Database
	jobChannel  chan []string
	stopChannel chan struct{}
	wg          sync.WaitGroup
}

//NewJobQueue function
func NewJobQueue(config *config.File, dnsbl *dnsbl.Dnsbl, db *db.Database) *JobQueue {
	jobQueue := JobQueue{
		dnsbl:       dnsbl,
		db:          db,
		jobChannel:  make(chan []string, config.JobQueue.QueueLength),
		stopChannel: make(chan struct{}),
		wg:          sync.WaitGroup{},
	}

	jobQueue.wg.Add(1)

	go jobQueue.worker()
	log.Println("job queue started")

	return &jobQueue
}

//Stop function
func (jq *JobQueue) Stop() bool {
	log.Println("stopping job queue")
	close(jq.stopChannel)
	ch := make(chan struct{})
	go func() {
		log.Println("waiting for job queue to stop")
		jq.wg.Wait()
		close(ch)
	}()
	select {
	case <-ch:
		log.Println("job queue stopped")
		return true
	case <-time.After(5 * time.Second):
		log.Println("timed out waiting for job queue to stop")
		return false
	}
}

//AddJob function
func (jq *JobQueue) AddJob(ipAddresses []string) bool {
	select {
	case jq.jobChannel <- ipAddresses:
		log.Printf("queued job for ip addresses: %+v\n", ipAddresses)
		return true
	default:
		log.Printf("queue busy - unable to queue job for ip addresses: %+v\n", ipAddresses)
		return false
	}
}

func (jq *JobQueue) worker() {
	defer jq.wg.Done()
	for {
		select {
		case <-jq.stopChannel:
			log.Println("job queue stopping")
			return

		case ipAddrs := <-jq.jobChannel:
			log.Printf("job queue begin processing job with ip addresses: %+v\n", ipAddrs)

			for _, ipAddr := range ipAddrs {
				resp := jq.dnsbl.Lookup(ipAddr)

				respCode := "NXDOMAIN"
				if resp.Responses[0].Resp != "" {
					respCode = resp.Responses[0].Resp
				}

				DNSBlockListRecord := model.DNSBlockListRecord{
					UUID:         uuid.New().String(),
					IPAddress:    ipAddr,
					ResponseCode: respCode,
				}

				jq.db.UpsertRecord(&DNSBlockListRecord)

				log.Printf("job queue completed processing for ip address: %s\n", ipAddr)
			}

			log.Printf("job queue completed processing job with ip addresses: %+v\n", ipAddrs)
		}
	}
}
