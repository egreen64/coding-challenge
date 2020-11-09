package config

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	//Disable logging
	log.SetOutput(ioutil.Discard)

	os.Setenv("GO_CONFIG", "../config.json")

	t.Run("get_config_success", func(t *testing.T) {

		config := GetConfig()
		require.NotEqual(t, nil, config)
	})
}
