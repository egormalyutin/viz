package logger

import (
	"fmt"
	"os"
	"strings"

	colors "github.com/fatih/color"
	bars "gopkg.in/cheggaaa/pb.v1"
)

func Fatal(text ...interface{}) {
	t := fmt.Sprint(text...)
	prefix := colors.RedString("ERR:")
	t = prefix + t
	fmt.Fprint(os.Stderr, t)
	os.Exit(1)
}

func Fatalf(text ...interface{}) {
	t := fmt.Sprintf(text[0].(string), text[1:]...)
	prefix := colors.RedString("ERR:")
	t = prefix + t
	fmt.Fprint(os.Stderr, t)
	os.Exit(1)
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
