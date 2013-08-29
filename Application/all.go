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
	cmd.Flag.IntVar(&CPUs, NUM_CPUS_FLAG, 0, "Max CPUS to use (runtime.GOMAXPROCS); 0 = all")
	for i, app := range cmd.Commands {
		app.Run = NewAppWrapCommand(app)
	}
	return commander
}

func WrapCommand(cmd *commander.Command, args []string) {
	numCPUs := cmd.Flag.Lookup(NUM_CPUS_FLAG)
	maxCPUs := runtime.NumCPU()
	if numCPUs > maxCPUs {
		log.Printf("Warning: Number of CPUs capped to all available (%d)", maxCPUs)
		numCPUs = 0
	}
	if numCPUs == 0 {
		numCPUs = maxCPUs
	}
	runtime.GOMAXPROCS(numCPUs)
}

func NewAppWrapCommand(command *commander.Command) *commander.Command {
	wrapped := func(cmd *commander.Command, args []string) {
		WrapperCommand(cmd, args)
		command.Run(cmd, args)
	}

	return wrapped
}
