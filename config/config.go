package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var config File
var readConfig = false

func init() {
	ReadConfig()
}

//ReadConfig function
func ReadConfig() *File {

	confPath := findConfig()
	var data []byte

	log.Printf("loading config: %s", confPath)
	r, err := os.Open(confPath)
	if err != nil {
		log.Fatalf(err.Error())
	}
	data, err = ioutil.ReadAll(r)
	if err != nil {
		log.Fatalf(err.Error())
	}
	r.Close()

	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &config
}

//GetConfig function
func GetConfig() *File {
	return &config
}

func findConfig() string {

	//go_config overrides any config
	if c := os.Getenv("go_config"); c != "" {
		return c
	}

	appenv := "dev"
	if _env := os.Getenv("APP_ENV"); _env != "" {
		appenv = _env
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pattern := filepath.Join(cwd, "config", "config.*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		panic(err)
	}

	configPath := ""
	for _, file := range files {
		base := filepath.Base(file)
		if strings.Contains(base, appenv) {
			configPath = file
		}
	}

	return configPath
}
