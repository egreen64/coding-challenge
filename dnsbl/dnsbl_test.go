package dnsbl

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/egreen64/codingchallenge/config"
	"github.com/stretchr/testify/require"
)

func TestDnsbl(t *testing.T) {
	//Disable logging
	log.SetOutput(ioutil.Discard)

	//Set location of config file
	os.Setenv("GO_CONFIG", "../config.json")

	//Get config file
	config := config.GetConfig()

	dnsbl := NewDnsbl(config)

	t.Run("new_dnsbl_success", func(t *testing.T) {

		newDnsbl := NewDnsbl(config)
		require.NotEqual(t, nil, newDnsbl)
	})

	t.Run("lookup_sucesss_127.0.0.2", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.2")

		require.NotEqual(t, nil, resp)
		require.Equal(t, true, resp.Responses[0].Listed)
		require.Equal(t, "127.0.0.2", resp.Responses[0].Resp)
	})
	t.Run("lookup_sucesss_127.0.0.3", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.3")

		require.NotEqual(t, nil, resp)
		require.Equal(t, true, resp.Responses[0].Listed)
		require.Equal(t, "127.0.0.3", resp.Responses[0].Resp)
	})
	t.Run("lookup_sucesss_127.0.0.4", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.4")

		require.NotEqual(t, nil, resp)
		require.Equal(t, true, resp.Responses[0].Listed)
		require.Equal(t, "127.0.0.4", resp.Responses[0].Resp)
	})
	t.Run("lookup_sucesss_127.0.0.9", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.9")

		require.NotEqual(t, nil, resp)
		require.Equal(t, true, resp.Responses[0].Listed)
		require.Equal(t, "127.0.0.2", resp.Responses[0].Resp)
	})
	t.Run("lookup_sucesss_127.0.0.10", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.10")

		require.NotEqual(t, nil, resp)
		require.Equal(t, true, resp.Responses[0].Listed)
		require.Equal(t, "127.0.0.10", resp.Responses[0].Resp)
	})
	t.Run("lookup_sucesss_127.0.0.11", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.11")

		require.NotEqual(t, nil, resp)
		require.Equal(t, true, resp.Responses[0].Listed)
		require.Equal(t, "127.0.0.11", resp.Responses[0].Resp)
	})
	t.Run("lookup_failure_not_listed_127.0.0.255", func(t *testing.T) {

		resp := dnsbl.Lookup("127.0.0.255")

		require.NotEqual(t, nil, resp)
		require.Equal(t, false, resp.Responses[0].Listed)
		require.Equal(t, "", resp.Responses[0].Resp)
	})
}
