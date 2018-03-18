package main

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/malyutinegor/viz/logger"
)

func Watch(filename string, changed chan bool, done chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.Print("write")
				}
			case err := <-watcher.Errors:
				logger.Fatal(fmt.Sprintf("Error while watching file %s: ", filename), err)
			}
		}
	}()

	err = watcher.Add(filename)
	if err != nil {
		logger.Fatal(err)
	}
	<-done
}

func WatchData(input string, output string, done chan bool) {
	changed := make(chan bool)
	go Watch(input, changed, done)
	for _ = range changed {
		Update(input, output)
	}
}
