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
	"github.com/egreen64/codingchallenge/auth"
	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/egreen64/codingchallenge/graph"
	"github.com/egreen64/codingchallenge/graph/generated"
	"github.com/go-chi/chi"
)

const defaultPort = "8080"

func main() {

	//Read config file
	config := config.GetConfig()

	//Initialize Database
	database := db.NewDatabase(config)

	//Instantiate DNS Blocklist instance
	dnsbl := dnsbl.NewDnsbl(config)
	resolver := graph.Resolver{
		Config:   config,
		Database: database,
		DNSBL:    dnsbl,
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

	router := chi.NewRouter()

	router.Use(auth.Middleware(config))

	//Instantiate graphql server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver}))

	//Initialize graphql handler functions
	router.Handle("/graphql", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	//Initialize listening port
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	//Start server on listening port
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)

	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		panic(err)
	}
}
