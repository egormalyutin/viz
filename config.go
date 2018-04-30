package main

import (
	"fmt"
	"io/ioutil"
	purl "net/url"

	"github.com/BurntSushi/toml"
)

type CSVConfig struct {
	File string
}

type Config struct {
	Provider         string
	Types            []string
	Headers          []string
	Animation        bool
	AnimateLastChunk bool
	Port             int
	CSV              CSVConfig
}

type ConfigJSON struct {
	Types            []string `json:"types"`
	Headers          []string `json:"headers"`
	Animation        bool     `json:"animation"`
	AnimateLastChunk bool     `json:"animateLastChunk"`
	WS               string   `json:"WS"`
}

var config Config

func ParseConfig() {
	configName := "viz.toml"

	data, err := ioutil.ReadFile(configName)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error when reading config file \"%s\": ", configName), err)
	}

	if err = toml.Unmarshal(data, &config); err != nil {
		logger.Fatal(fmt.Sprintf("Error when parsing config \"%s\": ", configName), err)
	}

	switch config.Provider {
	case "csv":
		Provider.Set("csv")
		Provider.InitCSV(config.CSV.File)

	default:
		logger.Fatalf("Not found provider \"%s\"!", config.Provider)
	}

	if config.Port == 0 {
		logger.Fatal("Port is not specified!")
	}
}

func getConfigJSON(url *purl.URL) ConfigJSON {
	logger.Print(url)
	return ConfigJSON{
		config.Types,
		config.Headers,
		config.Animation,
		config.AnimateLastChunk,
		"ws://" + url.Host + "/ws",
	}
}
