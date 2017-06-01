package app

import (
	"fmt"
	"log"

	"yap/nlp/format/lattice"
	"yap/nlp/format/mada"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func ArabMAConfigOut() {
	log.Println("Configuration")
	log.Printf("MADA Input:\t\t%s", input)
	log.Printf("Output:\t\t%s", outLatticeFile)
	log.Println()
}

func ArabMA(cmd *commander.Command, args []string) error {
	REQUIRED_FLAGS := []string{"mada", "out"}
	VerifyFlags(cmd, REQUIRED_FLAGS)

	ArabMAConfigOut()

	log.Println("Reading Morphological Analyis from MADA output file", input)
	madaLats, err := mada.ReadFile(input, limit)
	if err != nil {
		panic(fmt.Sprintf("Failed reading MADA file - %v", err))
	}
	log.Println("Converting to Internal Format")
	internalLats := mada.MADA2LatticeCorpus(madaLats)
	log.Println("Converting to Lattice Ouptut")
	lattices := lattice.Sentence2LatticeCorpus(internalLats, nil)
	log.Println("Writing to", outLatticeFile)
	err = lattice.WriteUDFile(outLatticeFile, lattices)
	if err != nil {
		panic(fmt.Sprintf("Failed writing to lattice file - %v", err))
	}
	return nil
}

func ArabMACmd() *commander.Command {
	cmd := &commander.Command{
		Run:       ArabMA,
		UsageLine: "arabma <file options> [arguments]",
		Short:     "convert MADA morphological analysis of arabic to UD",
		Long: `
convert MADA morphological analysis of arabic to UD

	$ ./yap arabma -mada <mada output file> -out <output file> [options]

`,
		Flag: *flag.NewFlagSet("arabma", flag.ExitOnError),
	}
	cmd.Flag.StringVar(&input, "mada", "", "Input MADA morphological analysis file")
	cmd.Flag.StringVar(&outLatticeFile, "out", "", "Output lattice file")
	cmd.Flag.BoolVar(&xliter8out, "xliter8out", false, "Transliterate output lattice file")
	cmd.Flag.IntVar(&limit, "limit", 0, "Limit input set")
	// cmd.Flag.BoolVar(&outJSON, "json", false, "Output using JSON")
	return cmd
}
