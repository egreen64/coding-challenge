package main

import (
	"io/ioutil"
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
	//Disable logging
	log.SetOutput(ioutil.Discard)

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

	var resp struct {
		Enqueue bool
	}

	mutation = `
		mutation {
			enqueue(ip: ["127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.9", "127.0.0.10", , "127.0.0.11", , "127.0.0.12"]) 
		}
	`
	c.Post(mutation, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
	require.Equal(t, true, resp.Enqueue)

	//Wait for blocklist jobs to complete
	time.Sleep(5 * time.Second)

	t.Run("authenticate_success", func(t *testing.T) {
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

	t.Run("authenticate_failure_invalid_credentials", func(t *testing.T) {
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
				enqueue(ip: ["127.0.0.13", "127.0.0.14", "127.0.0.15", "127.0.0.153", "127.0.0.163"]) 
			}
		`
		c.Post(mutation, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, true, resp.Enqueue)
	})

	t.Run("enqueue_success_update", func(t *testing.T) {

		var enqueueResp struct {
			Enqueue bool
		}

		mutation := `
			mutation {
				enqueue(ip: ["127.0.0.122"])
			}
		`
		c.Post(mutation, &enqueueResp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, true, enqueueResp.Enqueue)

		time.Sleep(1 * time.Second)

		var getResp1 struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.122")
				{
					uuid
					ip_address
					created_at
					updated_at
					response_code
				}
			}
		`
		c.Post(query, &getResp1, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.122", getResp1.GetIPDetails.IPAddress)
		require.Equal(t, "NXDOMAIN", getResp1.GetIPDetails.ResponseCode)

		time.Sleep(1 * time.Second)

		c.Post(mutation, &enqueueResp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, true, enqueueResp.Enqueue)

		time.Sleep(1 * time.Second)

		var getResp2 struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		c.Post(query, &getResp2, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.122", getResp2.GetIPDetails.IPAddress)
		require.Equal(t, "NXDOMAIN", getResp2.GetIPDetails.ResponseCode)

		require.Equal(t, getResp1.GetIPDetails.IPAddress, getResp2.GetIPDetails.IPAddress)
		require.Equal(t, getResp1.GetIPDetails.ResponseCode, getResp2.GetIPDetails.ResponseCode)
		require.Equal(t, getResp1.GetIPDetails.CreatedAt, getResp2.GetIPDetails.CreatedAt)
		require.NotEqual(t, getResp1.GetIPDetails.UpdatedAt, getResp2.GetIPDetails.UpdatedAt)
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

	t.Run("get_ip_details_success_127.0.0.2", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.2") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.2", resp.GetIPDetails.IPAddress)
		require.Equal(t, "127.0.0.2", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_success_127.0.0.3", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.3") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.3", resp.GetIPDetails.IPAddress)
		require.Equal(t, "127.0.0.3", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_success_127.0.0.4", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
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
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.4", resp.GetIPDetails.IPAddress)
		require.Equal(t, "127.0.0.4", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_success_127.0.0.9", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.9") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.9", resp.GetIPDetails.IPAddress)
		require.Equal(t, "127.0.0.2", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_success_127.0.0.10", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.10") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.10", resp.GetIPDetails.IPAddress)
		require.Equal(t, "127.0.0.10", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_success_127.0.0.11", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.11") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.11", resp.GetIPDetails.IPAddress)
		require.Equal(t, "127.0.0.11", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_success_NXDOMAIN", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.12") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.12", resp.GetIPDetails.IPAddress)
		require.Equal(t, "NXDOMAIN", resp.GetIPDetails.ResponseCode)
	})

	t.Run("get_ip_details_failure_no_auth_token", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
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
		err := c.Post(query, &resp)
		require.EqualError(t, err, `[{"message":"missing auth token","path":["getIPDetails"]}]`)
		require.Nil(t, resp.GetIPDetails)
	})

	t.Run("get_ip_details_failure_blocklist_not_found", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.92") 
				{
			  		uuid
			  		ip_address
			  		created_at
			  		updated_at
			  		response_code
				}
			}
		`
		c.Post(query, &resp, client.AddHeader("Authorization", authResp.Authenticate.BearerToken))
		require.Equal(t, "127.0.0.92", resp.GetIPDetails.IPAddress)
		require.Equal(t, "NXDOMAIN", resp.GetIPDetails.ResponseCode)
		require.Equal(t, "", resp.GetIPDetails.UUID)
		require.NotEmpty(t, resp.GetIPDetails.CreatedAt)
		require.NotEmpty(t, resp.GetIPDetails.UpdatedAt)
	})

	t.Run("get_ip_details_failure_invalid_ip_address", func(t *testing.T) {

		var resp struct {
			GetIPDetails *struct {
				UUID         string `json:"uuid"`
				CreatedAt    string `json:"created_at"`
				UpdatedAt    string `json:"updated_at"`
				ResponseCode string `json:"response_code"`
				IPAddress    string `json:"ip_address"`
			}
		}

		query := `
			{
				getIPDetails(ip:"127.0.0.444") 
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
		require.EqualError(t, err, `[{"message":"invalid IPV4 address: 127.0.0.444","path":["getIPDetails"]}]`)
		require.Nil(t, resp.GetIPDetails)
	})
}
