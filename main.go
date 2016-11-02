// +build !appengine

package main

import (
	_ "net/http/pprof"
	"yap/app"

	"github.com/gonuts/commander"

	"fmt"
	"os"
)

var cmd *commander.Command

func init() {
	cmd = app.AllCommands()
}

func exit(err error) {
	fmt.Printf("**error**: %v\n", err)
	os.Exit(1)
}

func main() {
	if err := cmd.Dispatch(os.Args[1:]); err != nil {
		exit(err)
	}

	return
}
