//+build ignore

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/GeertJohan/go.rice"

	evbus "github.com/asaskevich/EventBus"
	ws "github.com/gorilla/websocket"
)

var (
	Upgrader = ws.Upgrader{}

	Box       *rice.Box
	HTTPBox   *rice.HTTPBox
	HTTPFiles http.Handler

	StaticRoutes = []string{}

	WatchChannel     = make(chan bool)
	WatchDoneChannel = make(chan bool)

	stringTypes   string
	stringHeaders string

	bus = evbus.New()
)

type Request struct {
	Type  string `json:"type"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	ID    int    `json:"id"`
}

type LinesCountResponse struct {
	Type       string `json:"type"`
	LinesCount int    `json:"linesCount"`
}

type ReadResponse struct {
	Type  string `json:"type"`
	Lines string `json:"lines"`
	ID    int    `json:"id"`
}

type ErrorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	c, err := Upgrader.Upgrade(w, r, nil)

	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()

	wrt := func(text []byte) error {
		return c.WriteMessage(ws.TextMessage, text)
	}

	done := make(chan bool)

	linesCountWriter := func() {
		lines := Provider.Lines()
		resp := LinesCountResponse{"linesCount", lines}
		data, err := json.Marshal(resp)
		if err != nil {
			logger.Error("Error when marshaling LinesCount response:", err)
		} else {
			wrt(data)
		}
	}

	bus.Subscribe("provider:watch", linesCountWriter)
	defer bus.Unsubscribe("provider:watch", linesCountWriter)

	go func() {
		defer func() { done <- true }()

		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			handleError := func(err error) error {
				resp := ErrorResponse{"error", fmt.Sprint(err)}
				data, err2 := json.Marshal(resp)
				if err2 != nil {
					logger.Fatal("Error when generating error JSON:", err2)
				}
				logger.Println(string(data))
				return wrt(data)
			}

			request := &Request{}
			if err := json.Unmarshal(message, request); err != nil {
				if handleError(err) != nil {
					break
				}
			}

			switch request.Type {
			case "linesCount":
				lines := Provider.Lines()
				resp := LinesCountResponse{"linesCount", lines}
				data, err := json.Marshal(resp)
				if err != nil {
					logger.Error("Error when marshaling LinesCount response:", err)
				} else {
					wrt(data)
				}

			case "read":
				lines := Provider.Read(request.Start, request.End)
				str := strings.Join(lines, "\n")
				resp := ReadResponse{"read", str, request.ID}
				data, err := json.Marshal(resp)
				if err != nil {
					logger.Error("Error when marshaling Read response:", err)
				} else {
					wrt(data)
				}

			default:
				str := fmt.Sprintf("Not found command \"%s\"!", request.Type)
				resp := ErrorResponse{"error", str}
				data, err := json.Marshal(resp)
				if err != nil {
					logger.Error("Error when marshaling Error response:", err)
				} else {
					wrt(data)
				}
			}
		}
	}()

	<-done
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/config.js" {
		// generate config js
		logger.Print(r.URL)
		data, err := json.Marshal(getConfigJSON(r.URL))
		if err != nil {
			logger.Error("Error when trying to generate JS config: ", data)
		} else {
			data = []byte("var config = JSON.parse(`" + string(data) + "`);")
			w.Write(data)
		}
	} else {
		// serve static files
		HTTPFiles.ServeHTTP(w, r)
	}
}

func Serve() {
	// port
	port := config.Port
	portString := fmt.Sprintf(":%d", port)

	// box
	Box = rice.MustFindBox("dist")
	HTTPBox = Box.HTTPBox()
	HTTPFiles = http.FileServer(HTTPBox)

	logger.Printf("Listening on http://localhost%s", portString)

	// routes
	http.HandleFunc("/", HomeHandler)
	http.HandleFunc("/ws", WSHandler)

	done := make(chan bool)

	go func() {
		if err := http.ListenAndServe(portString, nil); err != nil {
			logger.Fatal(fmt.Sprintf("Error when listening on port %d: ", port), err)
		}
	}()

	go func() {
		ch := make(chan bool)
		go Provider.Watch(ch, done)
		for _ = range ch {
			bus.Publish("provider:watch")
		}
	}()

	<-done
	return
}
