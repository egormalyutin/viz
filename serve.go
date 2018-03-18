package main

import (
	"fmt"
	"net/http"

	"github.com/malyutinegor/viz/logger"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello!")
}

func Serve() {
	http.HandleFunc("/", MainHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal(err)
	}
}
