package main

import (
	"fmt"
	"os"
)

var logger = Logger{}

type Logger struct{}

func (_ Logger) Print(text ...interface{}) {
	t := fmt.Sprint(text...)
	fmt.Fprintln(os.Stdout, t)
}

func (l Logger) Println(text ...interface{}) {
	l.Print(text...)
}

func (_ Logger) Printf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	fmt.Fprintln(os.Stdout, t)
}

func (_ Logger) Fatal(text ...interface{}) {
	t := fmt.Sprint(text...)
	prefix := "FATAL: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
	os.Exit(1)
}

func (_ Logger) Fatalf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	prefix := "FATAL: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
	os.Exit(1)
}

func (_ Logger) Error(text ...interface{}) {
	t := fmt.Sprint(text...)
	prefix := "ERROR: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
}

func (_ Logger) Errorf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	prefix := "ERROR: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
}
