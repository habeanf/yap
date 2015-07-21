package app

import (
	"yap/nlp/format/lattice"
	"yap/nlp/format/raw"

	"yap/nlp/parser/ma"
	nlp "yap/nlp/types"
	// "yap/util"

	"fmt"
	"log"
	// "os"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	prefixFile, lexiconFile string
)

func HebMAConfigOut() {
	log.Println("Configuration")
	log.Printf("Heb Lexicon:\t\t%s", prefixFile)
	log.Printf("Heb Prefix:\t\t%s", lexiconFile)
	log.Printf("OOV Strategy:\t%v", "Const:NNP")
	log.Println()
	log.Printf("Raw Input:\t\t%s", inRawFile)
	log.Printf("Output:\t\t%s", outLatticeFile)
	log.Println()
}

func HebMA(cmd *commander.Command, args []string) {
	REQUIRED_FLAGS := []string{"prefix", "lexicon", "raw", "out"}
	VerifyFlags(cmd, REQUIRED_FLAGS)
	HebMAConfigOut()
	maData := new(ma.BGULex)
	// log.Println("Reading Morphological Analyzer BGU Prefixes")
	// maData.LoadPrefixes(prefixFile)
	log.Println("Reading Morphological Analyzer BGU Lexicon")
	maData.LoadLex(lexiconFile)
	log.Println()
	sents, err := raw.ReadFile(inRawFile)
	if err != nil {
		panic(fmt.Sprintf("Failed reading raw file - %v", err))
	}
	log.Println("Running Hebrew Morphological Analysis")
	lattices := make([]nlp.LatticeSentence, len(sents))
	stats := new(ma.AnalyzeStats)
	stats.Init()
	maData.Stats = stats
	for i, sent := range sents {
		lattices[i], _ = maData.Analyze(sent.Tokens())
	}
	log.Println("Analyzed", stats.TotalTokens, "occurences of", len(stats.UniqTokens), "unique tokens")
	log.Println("Encountered", stats.OOVTokens, "occurences of", len(stats.UniqOOVTokens), "unknown tokens")
	output := lattice.Sentence2LatticeCorpus(lattices)
	lattice.WriteFile(outLatticeFile, output)
	log.Println("Wrote", len(output), "lattices")
}

func HebMACmd() *commander.Command {
	cmd := &commander.Command{
		Run:       HebMA,
		UsageLine: "hebma <file options> [arguments]",
		Short:     "run lexicon-based morphological analyzer on raw input",
		Long: `
run lexicon-based morphological analyzer on raw input

	$ ./yap hebma -prefix <prefix file> -lexicon <lexicon file> -raw <raw file> -out <output file> [options]

`,
		Flag: *flag.NewFlagSet("ma", flag.ExitOnError),
	}
	cmd.Flag.StringVar(&prefixFile, "prefix", "", "Prefix file for morphological analyzer")
	cmd.Flag.StringVar(&lexiconFile, "lexicon", "", "Lexicon file for morphological analyzer")
	cmd.Flag.StringVar(&inRawFile, "raw", "", "Input raw (tokenized) file")
	cmd.Flag.StringVar(&outLatticeFile, "out", "", "Output lattice file")
	return cmd
}
