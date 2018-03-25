package main

import (
	"github.com/malyutinegor/viz/logger"
)

var (
	CurrentProvider = "csv"
	Provider        = &ProviderType{}
)

type ProviderType struct{}

//////////////////////////////////////////////////////

func (p *ProviderType) Init(cfg string) {
	switch CurrentProvider {
	case "csv":
		CSVProvider.Init(cfg)
	}
}

func (p *ProviderType) Read(start int, end int) []string {
	switch CurrentProvider {
	case "csv":
		return CSVProvider.Read(start, end)
	}
	return []string{}
}

func (p *ProviderType) Lines() int {
	switch CurrentProvider {
	case "csv":
		return CSVProvider.Lines()
	}
	return 0
}

//////////////////////////////////////////////////////

func (p *ProviderType) Set(name string) {
	if name != "csv" {
		logger.Fatalf("Provider %s not found!", name)
	}
	CurrentProvider = name
}
