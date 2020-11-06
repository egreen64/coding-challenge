package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/egreen64/codingchallenge/graph"
	"github.com/egreen64/codingchallenge/graph/generated"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	config := config.GetConfig()
	database := db.NewDatabase(config)

	//Instantiate DNS Blocklist instance
	dnsbl := dnsbl.NewDnsbl(config)
	resolver := graph.Resolver{
		Config:   config,
		Database: database,
		DNSBL:    dnsbl,
	}
	mainCtx, shutdown := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	//Handle termination signals
	go func() {
		for {
			s := <-sigChan
			log.Printf("codingchallege signal recieved '%s'", s)
			shutdown()
		}
	}()

	//Handle gracefule shutdown
	go func() {
		<-mainCtx.Done() //container going down
		log.Printf("container recieved TERM. Flushing.")
		database.CloseDatabase()
		os.Exit(1)
	}()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
