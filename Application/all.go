package Application

import (
	"chukuparser/Application/morphparse"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"os"

	"log"
	"net/http"
	"runtime"
	"runtime/pprof"
)

const (
	NUM_CPUS_FLAG = "cpus"
)

var (
	CPUs       int
	CPUProfile string
)

var AppCommands []*commander.Command = []*commander.Command{
	morphparse.MorphCmd(),
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
		app.Flag.StringVar(&CPUProfile, "cpuprofile", "", "write cpu profile to file")
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
	log.Printf("GOMAXPROCS:\t%d", CPUs)
	runtime.GOMAXPROCS(CPUs)

	// launch net server for profiling
	log.Println("Profiler interface:", "http://127.0.0.1:6060/debug/pprof")
	go func() {
		log.Println(http.ListenAndServe("127.0.0.1:6060", nil))
	}()
}

func NewAppWrapCommand(f func(cmd *commander.Command, args []string)) func(cmd *commander.Command, args []string) {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	wrapped := func(cmd *commander.Command, args []string) {
		InitCommand(cmd, args)
		if CPUProfile != "" {
			f, err := os.Create(CPUProfile)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("Writing profiling info to", CPUProfile)
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
		log.Println()
		f(cmd, args)
	}

	return wrapped
}
