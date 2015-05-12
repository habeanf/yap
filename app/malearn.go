package app

import (
	// "yap/nlp/format/lattice"

	// nlp "yap/nlp/types"
	// "yap/util"

	// "fmt"
	"log"
	// "os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var latfile string

func MALearnConfigOut() {
	log.Println("Configuration")
}

func MALearn(cmd *commander.Command, args []string) {
	REQUIRED_FLAGS := []string{"lattice"}

	VerifyFlags(cmd, REQUIRED_FLAGS)
	// RegisterTypes()

	MALearnConfigOut()
	log.Println("Bla")
}

func MALearnCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       MALearn,
		UsageLine: "malearn <file options> [arguments]",
		Short:     "generate a data-driven morphological analysis dictionary for a set of files",
		Long: `
generate a data-driven morphological analysis dictionary for a set of files

	$ ./yap malearn -lattice [options]

`,
		Flag: *flag.NewFlagSet("malearn", flag.ExitOnError),
	}
	cmd.Flag.StringVar(&latfile, "lattice", "", "Lattice-format input file")
	return cmd
}
