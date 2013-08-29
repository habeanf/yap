package main

import (
	"chukuparser/Application"
	"github.com/gonuts/commander"
	_ "net/http/pprof"
)

var cmd *commander.Commander

func init() {
	cmd = Application.AllCommands()
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
