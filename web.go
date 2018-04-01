package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"

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

	jsTemplate *template.Template

	slashRG  = regexp.MustCompile(`\\`)
	quotesRG = regexp.MustCompile(`"`)

	stringTypes   string
	stringHeaders string

	bus = evbus.New()
)

//easyjson:json
type Request struct {
	Type  string `json:"type"`
	Start int    `json:"start"`
	End   int    `json:"end"`
	ID    int    `json:"id"`
}

//easyjson:json
type LinesCountResponse struct {
	Type       string `json:"type"`
	LinesCount int    `json:"linesCount"`
}

//easyjson:json
type ReadResponse struct {
	Type  string `json:"type"`
	Lines string `json:"lines"`
	ID    int    `json:"id"`
}

//easyjson:json
type ErrorResponse struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

type JSTemplate struct {
	WS      string
	Types   string
	Headers string
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
		data, err := resp.MarshalJSON()
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
				data, err2 := resp.MarshalJSON()
				if err2 != nil {
					logger.Fatal("Error when generating error JSON:", err2)
				}
				logger.Println(string(data))
				return wrt(data)
			}

			request := &Request{}
			if err := request.UnmarshalJSON(message); err != nil {
				if handleError(err) != nil {
					break
				}
			}

			switch request.Type {
			case "linesCount":
				lines := Provider.Lines()
				resp := LinesCountResponse{"linesCount", lines}
				data, err := resp.MarshalJSON()
				if err != nil {
					logger.Error("Error when marshaling LinesCount response:", err)
				} else {
					wrt(data)
				}

			case "read":
				lines := Provider.Read(request.Start, request.End)
				str := strings.Join(lines, "\n")
				resp := ReadResponse{"read", str, request.ID}
				data, err := resp.MarshalJSON()
				if err != nil {
					logger.Error("Error when marshaling Read response:", err)
				} else {
					wrt(data)
				}

			default:
				str := fmt.Sprintf("Not found command \"%s\"!", request.Type)
				resp := ErrorResponse{"error", str}
				data, err := resp.MarshalJSON()
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

func runTemplate(w http.ResponseWriter, host string) {
	tmpl := JSTemplate{
		quotesRG.ReplaceAllString(slashRG.ReplaceAllString("ws://"+host+"/ws", "\\\\"), "\\\""),
		quotesRG.ReplaceAllString(slashRG.ReplaceAllString(stringTypes, "\\\\"), "\\\""),
		quotesRG.ReplaceAllString(slashRG.ReplaceAllString(stringHeaders, "\\\\"), "\\\""),
	}

	jsTemplate.Execute(w, tmpl)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/main.js" {
		// serve main script template
		CompileTemplate()

		runTemplate(w, r.Host)

	} else {
		// serve static files
		HTTPFiles.ServeHTTP(w, r)
	}
}

func CompileTemplate() {
	templateString, err := Box.String("main.js")
	if err != nil {
		logger.Fatal("Error when trying to get main.js: ", err)
	}
	jsTemplate, err = template.New("jsresponse").Parse(templateString)
	if err != nil {
		logger.Fatal("Error when trying to compile main.js template: ", err)
	}
}

func Serve() {
	byteTypes, err := json.Marshal(config.Types)
	if err != nil {
		logger.Fatal("Error when marshaling Types response: ", err)
	}
	stringTypes = string(byteTypes)

	byteHeaders, err := json.Marshal(config.Headers)
	if err != nil {
		logger.Fatal("Error when marshaling Headers response: ", err)
	}
	stringHeaders = string(byteHeaders)

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
		err = http.ListenAndServe(portString, nil)

		if err != nil {
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
