package main

import (
	"github.com/malyutinegor/viz/logger"
)

func main() {
	AnalyzeInitial("data.csv", "output.html")

	go Serve()
	WatchData("data.csv", "output.html", make(chan bool))

}
