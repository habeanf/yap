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

func main() {
	err := cmd.Flag.Parse(os.Args[1:])
	if err != nil {
		fmt.Printf("**err**: %v\n", err)
		os.Exit(1)
	}

	args := cmd.Flag.Args()
	err = cmd.Run(args)
	if err != nil {
		fmt.Printf("**err**: %v\n", err)
		os.Exit(1)
	}

	return
}
