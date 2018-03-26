package logger

import (
	"fmt"
	"os"
)

func Print(text ...interface{}) {
	t := fmt.Sprint(text...)
	fmt.Fprintln(os.Stdout, t)
}

func Println(text ...interface{}) {
	Print(text...)
}

func Printf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	fmt.Fprintln(os.Stdout, t)
}

func Fatal(text ...interface{}) {
	t := fmt.Sprint(text...)
	prefix := "FATAL: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
	os.Exit(1)
}

func Fatalf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	prefix := "FATAL: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
	os.Exit(1)
}

func Error(text ...interface{}) {
	t := fmt.Sprint(text...)
	prefix := "ERROR: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
}

func Errorf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	prefix := "ERROR: "
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
}
