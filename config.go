package main

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type CSVConfig struct {
	File string
}

type Config struct {
	Provider string
	Types    []string
	Headers  []string
	Port     int
	CSV      CSVConfig
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
