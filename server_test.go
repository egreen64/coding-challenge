package main

import (
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"

	"github.com/egreen64/codingchallenge/auth"
	"github.com/egreen64/codingchallenge/config"
	"github.com/egreen64/codingchallenge/db"
	"github.com/egreen64/codingchallenge/dnsbl"
	"github.com/egreen64/codingchallenge/graph"
	"github.com/egreen64/codingchallenge/graph/generated"
	"github.com/egreen64/codingchallenge/jobqueue"
)

func TestCodingChallenge(t *testing.T) {
	//Read config file
	config := config.GetConfig()

	//Initialize databse
	database := db.NewDatabase(config)

	//Instantiate DNS Blocklist instance
	dnsbl := dnsbl.NewDnsbl(config)

	//Instantiage job queue
	jobQueue := jobqueue.NewJobQueue(dnsbl, database)

	//Initialize resolver
	resolver := graph.Resolver{
		Config:   config,
		Database: database,
		DNSBL:    dnsbl,
		JobQueue: jobQueue,
	}

	//Instantiate router
	router := chi.NewRouter()

	//Use authentication middleware
	router.Use(auth.Middleware(config))

	//Instantiate graphql server
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolver}))

	//Instantiate graphql client
	c := client.New(router)

	//Initialize graphql handler functions
	router.Handle("/", srv)

	go func() {
		port := "8080"
		//Start server on listening port
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			panic(err)
		}
	}()

	var authResp struct {
		Authenticate struct {
			BearerToken string `json:"bearerToken"`
		}
	}

	mutation := `
		mutation { 
			authenticate(username: "secureworks", password: "supersecret") 
			{ bearerToken } 
		}
	`
	c.MustPost(mutation, &authResp)

	require.Equal(t, "Bearer", strings.Split(authResp.Authenticate.BearerToken, " ")[0])

	t.Run("obtain_authentication_token_success", func(t *testing.T) {
		var resp struct {
			Authenticate struct {
				BearerToken string `json:"bearerToken"`
			}
		}

		mutation := `
			mutation { 
				authenticate(username: "secureworks", password: "supersecret") 
				{ 
					bearerToken 
				} 
			}
		`

		c.MustPost(mutation, &resp)

		require.Equal(t, "Bearer", strings.Split(resp.Authenticate.BearerToken, " ")[0])
	})

	t.Run("obtain_authentication_token_invalid_credentials", func(t *testing.T) {
		var resp struct {
			Authenticate struct {
				BearerToken string `json:"bearerToken"`
			}
		}

		mutation := `
			mutation { 
				authenticate(username: "bozo", password: "clown") 
				{ 
					bearerToken 
				} 
			}
		`
		err := c.Post(mutation, &resp)
		require.EqualError(t, err, `[{"message":"invalid credentials","path":["authenticate"]}]`)
		require.Equal(t, "", resp.Authenticate.BearerToken)
	})

	t.Run("enqueue_success", func(t *testing.T) {
		var resp struct {
			Enqueue bool
		}

		mutation := `
			mutation { 
				enqueue(ip: ["127.0.0.3", "127.0.0.4", "127.0.0.5", "127.0.0.153", "127.0.0.163"]) 
			}
		`
		c.Post(mutation, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, true, resp.Enqueue)
	})

	t.Run("enqueue_failure_no_auth_token", func(t *testing.T) {
		var resp struct {
			Enqueue bool
		}

		mutation := `
			mutation { 
				enqueue(ip: ["127.0.0.3", "127.0.0.4", "127.0.0.5", "127.0.0.153", "127.0.0.163"]) 
			}
		`

		err := c.Post(mutation, &resp)
		require.EqualError(t, err, `[{"message":"missing auth token","path":["enqueue"]}]`)
		require.Equal(t, false, resp.Enqueue)
	})

	t.Run("enqueue_failure_invalid_ip_address", func(t *testing.T) {
		var resp struct {
			Enqueue bool
		}

		mutation := `
			mutation { 
				enqueue(ip: ["127.0.0.256"]) 
			}
		`

		err := c.Post(mutation, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.EqualError(t, err, `[{"message":"invalid IPV4 address: 127.0.0.256","path":["enqueue"]},{"message":"validation error(s)","path":["enqueue"]}]`)
		require.Equal(t, false, resp.Enqueue)
	})

	t.Run("enqueue_failure_invalid_ip_addresses", func(t *testing.T) {
		var resp struct {
			Enqueue bool
		}

		mutation := `
			mutation { 
				enqueue(ip: ["127.0.0.256", "127.0.0.257"]) 
			}
		`

		err := c.Post(mutation, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.EqualError(t, err, `[{"message":"invalid IPV4 address: 127.0.0.256","path":["enqueue"]},{"message":"invalid IPV4 address: 127.0.0.257","path":["enqueue"]},{"message":"validation error(s)","path":["enqueue"]}]`)
		require.Equal(t, false, resp.Enqueue)
	})

	t.Run("get_ip_details_success", func(t *testing.T) {

		var resp struct {
			GetIPDetails struct {
				UUID         string    `json:"uuid"`
				CreatedAt    time.Time `json:"created_at"`
				UpdatedAt    time.Time `json:"updated_at"`
				ResponseCode string    `json:"response_code"`
				IPAddress    string    `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.4") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		err := c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		log.Printf("%s\n", err)
		require.Equal(t, "127.0.0.4", resp.GetIPDetails.ResponseCode)
	})
}
