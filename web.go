package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/malyutinegor/viz/logger"

	"github.com/GeertJohan/go.rice"

	ws "github.com/gorilla/websocket"
)

var (
	Upgrader     = ws.Upgrader{}
	Box          *rice.Box
	HTTPBox      *rice.HTTPBox
	HTTPFiles    http.Handler
	StaticRoutes = make([]string, 0)
)

func WSHandler(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrader.Upgrade(w, r, nil)

	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		} else {
			str := string(message)

			// send lines count to client
			if str == "lines" {
				linesCount := []byte(fmt.Sprintf("lines:%d", Provider.Lines()))
				err = c.WriteMessage(ws.TextMessage, linesCount)
				if err != nil {
					logger.Error("WebSocket write err-1 message error: ", err)
				}
				continue
			}

			nums := strings.Split(str, ":")
			// nums length must be 2
			if len(nums) < 2 {
				err = c.WriteMessage(ws.TextMessage, []byte("err-1"))
				if err != nil {
					logger.Error("WebSocket write err-1 message error: ", err)
					continue
				}
			}
			// scan start number
			starts, err := strconv.ParseInt(nums[0], 10, 64)
			if err != nil {
				logger.Error("WebSocket write err-1 message error: ", err)
				continue
			}

			// scan end number
			ends, err := strconv.ParseInt(nums[1], 10, 64)
			if err != nil {
				logger.Error("WebSocket write err-1 message error: ", err)
				continue
			}

			// read lines from provider
			start := int(starts)
			end := int(ends)

			lines := Provider.Read(start, end)

			response := strings.Join(lines, "\n")

			responseBytes := []byte(response)
			err = c.WriteMessage(ws.TextMessage, responseBytes)
			if err != nil {
				logger.Error("WebSocket write message error: ", err)
			}
		}
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/main.js" {
		// serve main script template

		templateString, err := Box.String("main.js")
		if err != nil {
			logger.Fatal("Error when trying to get main.js: ", err)
		}
		tmplMessage, err := template.New("jsresponse").Parse(templateString)
		if err != nil {
			logger.Fatal("Error when trying to compile main.js template: ", err)
		}

		addr := "ws://" + r.Host + "/ws"
		tmplMessage.Execute(w, addr)

	} else {
		// serve static files
		HTTPFiles.ServeHTTP(w, r)
	}
}

func Serve() {
	// port
	port := 8080
	portString := fmt.Sprintf(":%d", port)

	// box
	Box = rice.MustFindBox("dist")
	HTTPBox = Box.HTTPBox()
	HTTPFiles = http.FileServer(HTTPBox)

	logger.Printf("Listening on http://localhost%s", portString)

	// routes
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/ws", WSHandler)

	err := http.ListenAndServe(portString, nil)

	if err != nil {
		logger.Fatal(fmt.Sprintf("Error when listening on port %d: ", port), err)
	}
}
