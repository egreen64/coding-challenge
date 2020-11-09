package utils

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUtils(t *testing.T) {
	//Disable logging
	log.SetOutput(ioutil.Discard)

	t.Run("file_exists_success", func(t *testing.T) {

		resp := FileExists("../config.json")
		require.Equal(t, true, resp)
	})

	t.Run("file_exists_failure", func(t *testing.T) {

		resp := FileExists("../configs.json")
		require.Equal(t, false, resp)
	})

	t.Run("valid_ipv4_address_success", func(t *testing.T) {

		resp := IsValidIPV4Address("127.0.0.1")
		require.Equal(t, true, resp)
	})

	t.Run("valid_ipv4_address_failure", func(t *testing.T) {

		resp := IsValidIPV4Address("127.0.0.256")
		require.Equal(t, false, resp)
	})

}
