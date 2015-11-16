// +build !appengine

package main

import (
	"github.com/gonuts/commander"
	_ "net/http/pprof"
	"yap/app"

	"fmt"
	"os"
)

var cmd *commander.Commander

func init() {
	cmd = app.AllCommands()
}

func exit(err error) {
	fmt.Printf("**error**: %v\n", err)
	os.Exit(1)
}

func main() {
	if err := cmd.Flag.Parse(os.Args[1:]); err != nil {
		exit(err)
	}

	args := cmd.Flag.Args()
	if err := cmd.Run(args); err != nil {
		exit(err)
	}

	return
}
