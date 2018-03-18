package main

import (
	"encoding/csv"
	// "fmt"
	"bufio"
	"io"
	"os"

	"github.com/malyutinegor/viz/logger"
)

func CSVChannel(file *os.File, channel chan []string, progress chan int) {
	reader := csv.NewReader(file)
	reader.Comma = ';'

	lines := 0

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			logger.Fatal("Error when parsing CSV file:", err)
			return
		}

		lines++

		channel <- record
		progress <- lines
	}

	close(channel)
	close(progress)
}

func RenderCSV(ch chan []string, output string) {
	// create output writer
	out, err := os.Create(output)
	if err != nil {
		logger.Fatal("Error when creating output HTML file:", err)
	}
	defer out.Close()

	// write to output
	wr := func(str string) {
		out.Write([]byte(str))
	}

	wr("<table>\n")

	for record := range ch {
		wr("<tr>")

		for _, r := range record {
			wr("<td>" + r + "</td>")
		}

		wr("</tr>\n")
	}

	wr("</table>")
}

func lineCount(filename string) (int64, error) {
	lc := int64(0)

	f, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		lc++
	}

	return lc, s.Err()
}

func AnalyzeInitial(input string, output string) {
	// create input reader
	count, err := lineCount(input)
	if err != nil {
		logger.Fatal("Error when getting count of lines in CSV file:", err)
	}

	file, err := os.Open(input)
	if err != nil {
		logger.Fatal("Error when reading CSV file:", err)
	}
	defer file.Close()

	// init CSV
	ch := make(chan []string)
	pch := make(chan int)
	go CSVChannel(file, ch, pch)
	go RenderCSV(ch, output)

	bar := logger.Bar("Analyzing initial CSV file...", int(count))

	for p := range pch {
		bar.Set(p)
	}

	bar.FinishPrint("Success!")
}
