package logger

import (
	"fmt"
	"os"
	"strings"

	colors "github.com/fatih/color"
	bars "gopkg.in/cheggaaa/pb.v1"
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
	prefix := colors.RedString("FATAL: ")
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
	os.Exit(1)
}

func Fatalf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	prefix := colors.RedString("FATAL:")
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
	os.Exit(1)
}

func Error(text ...interface{}) {
	t := fmt.Sprint(text...)
	prefix := colors.RedString("ERROR: ")
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
}

func Errorf(templ string, text ...interface{}) {
	t := fmt.Sprintf(templ, text...)
	prefix := colors.RedString("ERROR:")
	t = prefix + t
	fmt.Fprintln(os.Stderr, t)
}

func Bar(prefix string, count int) *bars.ProgressBar {
	bar := bars.New(count).
		Prefix(prefix).
		Format(strings.Join([]string{
			"[",
			"#",
			" ",
			" ",
			"]",
		}, "\x00")).
		SetMaxWidth(100)

	bar.Start()
	return bar
}
