package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/malyutinegor/viz/logger"
)

type CSVProviderType struct {
	filename string
	file     *os.File
}

// func CSVToHTML(c []string) string {
// 	str := strings.Join(c, "\n")
// 	reader := csv.NewReader(strings.NewReader(str))
// 	reader.Comma = ';'
// 	result := ""

// 	for {
// 		record, err := reader.Read()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			logger.Fatal("Error when trying to read CSV line: ", err)
// 		}

// 		lr := "<tr>"
// 		for _, prop := range record {
// 			lr += "<td>" + prop + "</td>"
// 		}
// 		lr += "</tr>"

// 		result += lr + "\n"
// 	}

// 	return result
// }

var CSVProvider = &CSVProviderType{}

// init CSV provider
func (p *CSVProviderType) Init(cfg string) {
	var err error
	p.file, err = os.Open(cfg)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error when opening input file %s: ", cfg), err)
	}
}

func (p *CSVProviderType) Lines() int {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := p.file.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count

		case err != nil:
			logger.Fatal(fmt.Sprintf("Error when trying to get number of lines in file %s: ", p.filename), err)
		}
	}
}

// read lines from start to (end - 1)
func (p *CSVProviderType) Read(start int, end int) []string {
	acc := make([]string, 0)

	p.file.Seek(0, 0)
	scanner := bufio.NewScanner(p.file)
	current := 0
	for scanner.Scan() {
		if current >= end {
			break
		}
		if current >= start {
			acc = append(acc, scanner.Text())
		}

		current++
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal(fmt.Sprintf("Error when reading lines from file %s: ", p.filename), err)
	}

	return acc
}

func (p *CSVProviderType) Watch(ch chan bool, done chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal("Error when creating fsnotify watcher: ", err)
	}
	defer watcher.Close()

	cl := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					ch <- true
				}
			case err := <-watcher.Errors:
				logger.Fatal(fmt.Sprintf("Error when watching file \"%s\": ", p.filename), err)
			case <-cl:
				return
			}
		}
	}()

	err = watcher.Add(p.filename)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Error when adding file \"%s\" to file watcher: ", p.filename), err)
	}

	<-done
	cl <- true
}
