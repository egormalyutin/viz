package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/malyutinegor/viz/logger"

	"github.com/GeertJohan/go.rice"

	ws "github.com/gorilla/websocket"
)

var (
	Upgrader = ws.Upgrader{}

	Box       *rice.Box
	HTTPBox   *rice.HTTPBox
	HTTPFiles http.Handler

	StaticRoutes = make([]string, 0)

	WatchChannel     = make(chan bool)
	WatchDoneChannel = make(chan bool)
)

type Request struct {
	Type  string
	Start int
	End   int
	ID    int
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

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			break
		}

		wrt := func(text []byte) error {
			return c.WriteMessage(ws.TextMessage, text)
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

		request := Request{}
		if err := json.Unmarshal(message, &request); err != nil {
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

	go func() {
		err := http.ListenAndServe(portString, nil)

		if err != nil {
			logger.Fatal(fmt.Sprintf("Error when listening on port %d: ", port), err)
		}
	}()

	go Provider.Watch(WatchChannel, WatchDoneChannel)
	for _ = range WatchChannel {
		logger.Print("Changed")
	}
}
