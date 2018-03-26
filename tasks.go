// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"gopkg.in/alecthomas/kingpin.v2"
)

////////////////////////////////////////////////////////

func hr(err ...interface{}) {
	fmt.Fprint(os.Stderr, err...)
	fmt.Print("\n")
	os.Exit(1)
}

func hrf(e1 string, e2 ...interface{}) {
	fmt.Fprintf(os.Stderr, e1, e2...)
	fmt.Print("\n")
	os.Exit(1)
}

func log(data ...interface{}) {
	fmt.Println(data...)
}

func logf(e1 string, e2 ...interface{}) {
	fmt.Printf(e1, e2...)
	fmt.Print("\n")
}

////////////////////////////////////////////////////////

func run(command string, commands ...string) {
	outSlice := make([]interface{}, 0)
	outSlice = append(outSlice, ">")
	outSlice = append(outSlice, command)

	for _, c := range commands {
		outSlice = append(outSlice, interface{}(c))
	}

	fmt.Println(outSlice...)

	cmd := exec.Command(command, commands...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		hr(err)
	}
}

func checkPath(pth string, name string) {
	_, err := exec.LookPath(pth)
	if err == nil {
		logf("%s found...", name)
	} else {
		hrf("Please, install %s and return!", name)
	}
}

////////////////////////////////////////////////////////

func PrepareTask() {
	checkPath("node", "Node.JS")
	checkPath("go", "Go")
	checkPath("npm", "NPM")

	_, err := exec.LookPath("gulp")
	if err == nil {
		log("Gulp found...")
	} else {
		log("Gulp not found! Installing it...")
		if runtime.GOOS == "linux" {
			run("sudo", "npm", "i", "-g", "gulp-cli")
		} else {
			run("npm", "i", "-g", "gulp-cli")
		}
	}

	log("Installing NPM dependencies...")
	run("npm", "i")

	log("Installing Go dependencies...")
	run("go", "get", "./...", "github.com/mitchellh/gox", "github.com/GeertJohan/go.rice", "github.com/GeertJohan/go.rice/rice")

	log("Ready for development!")
}

func WatchTask() {
	run("gulp", "watch")
}

func RunTask() {
	run("go", "run", "config.go", "csv_provider.go", "providers.go", "web.go", "viz.go")
}

func ProductionTask() {
	run("gulp", "build", "--production")
	run("rice", "embed-go")
	run("gox", "-output=build/{{.Dir}}_{{.OS}}_{{.Arch}}")
}

////////////////////////////////////////////////////////

func main() {
	app := kingpin.New("tasks", "Task runner for viz")

	app.Command("prepare", "Prepare task").Action(func(c *kingpin.ParseContext) error {
		PrepareTask()
		return nil
	})

	app.Command("watch", "Watch task").Action(func(c *kingpin.ParseContext) error {
		WatchTask()
		return nil
	})

	app.Command("production", "Production build task").Action(func(c *kingpin.ParseContext) error {
		ProductionTask()
		return nil
	})

	app.Command("run", "Run task").Action(func(c *kingpin.ParseContext) error {
		RunTask()
		return nil
	})

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
