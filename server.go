package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/egreen64/codingchallenge/auth"
	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/egreen64/codingchallenge/graph"
	"github.com/egreen64/codingchallenge/graph/generated"
	"github.com/egreen64/codingchallenge/jobqueue"
	"github.com/go-chi/chi"
)

func main() {

	//Read config file
	config := config.GetConfig()

	//Initialize logging
	logFile, err := os.OpenFile(config.Logger.LogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	//Initialize databse
	database := db.NewDatabase(config)

	//Instantiate DNS Blocklist instance
	dnsbl := dnsbl.NewDnsbl(config)

	//Instantiage job queue
	jobQueue := jobqueue.NewJobQueue(config, dnsbl, database)

	//Initialize resolver
	resolver := graph.Resolver{
		Config:   config,
		Database: database,
		DNSBL:    dnsbl,
		JobQueue: jobQueue,
	}

	//Obtain main context
	mainCtx, shutdown := context.WithCancel(context.Background())

	//Create termination signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	//Handle termination signals
	go func() {
		for {
			s := <-sigChan
			log.Printf("%s recieved signal '%s'", os.Args[0], s)
			shutdown()
		}
	}()

	//Handle graceful shutdown
	go func() {
		<-mainCtx.Done() //container going down
		log.Printf("%s recieved termination signal. shutting down...", os.Args[0])
		jobQueue.Stop()
		database.CloseDatabase()
		log.Printf("%s shutdown complete", os.Args[0])
		os.Exit(1)
	}()

	//Instantiate router
	router := chi.NewRouter()

	//Use authentication middleware
	router.Use(auth.Middleware())

	//Instantiate graphql server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver}))

	//Initialize graphql handler functions
	router.Handle("/graphql", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/liveness", LivenessCheck)
	router.HandleFunc("/readiness", ReadinessCheck)

	//Initialize listening port
	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(config.Server.ListeningPort)
	}

	//Start server on listening port
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}

//LivenessCheck function
func LivenessCheck(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	res.Header().Add("Access-Control-Allow-Origin", "*")
	res.Header().Add("Access-Control-Allow-Methods", "*")
	res.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	res.Header().Add("Access-Control-Max-Age", "3600")
	fmt.Fprintf(res, "ok")
}

//ReadinessCheck function
func ReadinessCheck(res http.ResponseWriter, req *http.Request) {
	res.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
	res.Header().Add("Access-Control-Allow-Origin", "*")
	res.Header().Add("Access-Control-Allow-Methods", "*")
	res.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	res.Header().Add("Access-Control-Max-Age", "3600")
	fmt.Fprintf(res, "ok")
}
