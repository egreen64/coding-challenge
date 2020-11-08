package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/egreen64/codingchallenge/utils"
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
	if c := os.Getenv("GO_CONFIG"); c != "" {
		return c
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	configPath := filepath.Join(cwd, "config.json")

	if !utils.FileExists(configPath) {
		err := errors.New("config.json file not found")
		panic(err)
	}

	return configPath
}

//File struct
type File struct {
	Logger   Logger   `json:"logger"`
	Database Database `json:"db"`
	Dnsbl    Dnsbl    `json:"dnsbl"`
	Auth     Auth     `json:"auth"`
}

//Logger type
type Logger struct {
	LogFileName string `json:"log_file_name"`
}

//Database type
type Database struct {
	Persist bool   `json:"persist"`
	DbType  string `json:"db_type"`
	DbPath  string `json:"db_path"`
}

//Dnsbl type
type Dnsbl struct {
	BlocklistDomains []string `json:"blocklist_domains"`
}

//Auth type
type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
