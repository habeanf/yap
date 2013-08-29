package Application

import (
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"os"

	"log"
	"runtime"
)

const (
	NUM_CPUS_FLAG = "cpus"
)

var (
	CPUs int
)

var AppCommands []*commander.Command = []*commander.Command{
	MorphCmd(),
}

func AllCommands() *commander.Commander {
	cmd := &commander.Commander{
		Name:     os.Args[0],
		Commands: AppCommands,
		Flag:     flag.NewFlagSet("app", flag.ExitOnError),
	}
	for _, app := range cmd.Commands {
		app.Run = NewAppWrapCommand(app.Run)
		app.Flag.IntVar(&CPUs, NUM_CPUS_FLAG, 0, "Max CPUS to use (runtime.GOMAXPROCS); 0 = all")
	}
	return cmd
}

func InitCommand(cmd *commander.Command, args []string) {
	maxCPUs := runtime.NumCPU()
	if CPUs > maxCPUs {
		log.Printf("Warning: Number of CPUs capped to all available (%d)", maxCPUs)
		CPUs = 0
	}
	if CPUs == 0 {
		CPUs = maxCPUs
	}
	runtime.GOMAXPROCS(CPUs)
}

func NewAppWrapCommand(f func(cmd *commander.Command, args []string)) func(cmd *commander.Command, args []string) {
	wrapped := func(cmd *commander.Command, args []string) {
		InitCommand(cmd, args)
		f(cmd, args)
	}

	return wrapped
}
