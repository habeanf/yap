package app

import (
	"fmt"
	"os"
	"yap/nlp/format/conllu"
	"yap/nlp/format/lattice"
	"yap/nlp/parser/disambig"

	"log"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

var (
	FUSEuseConLLU bool
)

func FuseConfigOut() {
	log.Println("Configuration")
	log.Printf("Use CoNLL-U:\t\t%v", FUSEuseConLLU)
	log.Printf("Limit:\t\t%v", limit)
	log.Println()
	log.Println("Data")
	log.Printf("Test file  (ambig.  lattice):\t%s", input)
	if !VerifyExists(input) {
		return
	}
	if len(inputGold) > 0 {
		log.Printf("Test file  (disambig.  lattice):\t%s", inputGold)
		if !VerifyExists(inputGold) {
			return
		}
	}
	log.Printf("Out (disamb.) file:\t\t\t%s", outMap)
}

func Fuse(cmd *commander.Command, args []string) error {
	REQUIRED_FLAGS := []string{"l", "d", "o"}

	VerifyFlags(cmd, REQUIRED_FLAGS)

	FuseConfigOut()
	SetupEnum([]string{})

	var (
		lAmb  chan lattice.Lattice
		lAmbE error
	)
	if FUSEuseConLLU {
		log.Println("Amb. Lat:\tReading ambiguous conllul lattices from", input)
		lAmb, lAmbE = lattice.StreamULFile(input, limit)
	} else {
		log.Println("Amb. Lat:\tReading ambiguous conllu lattices from", input)
		lAmb, lAmbE = lattice.StreamFile(input, limit)
	}
	if lAmbE != nil {
		log.Println(lAmbE)
		return lAmbE
	}
	log.Println("Streaming to ambiguous lattice conversion")
	predAmbLatStream := lattice.Lattice2SentenceStream(lAmb, EWord, EPOS, EWPOS, EMorphProp, EMHost, EMSuffix)

	log.Println("Streaming disambiguation lattice from", inputGold)
	var lDis chan interface{}
	if FUSEuseConLLU {
		conlluStream, err := conllu.ReadFileAsStream(inputGold, limit)
		if err != nil {
			log.Println(err)
			return err
		}
		lDis = conllu.ConllU2MorphGraphStream(conlluStream, EWord, EPOS, EWPOS, ERel, EMorphProp, EMHost, EMSuffix)
	} else {
		panic("Unsupported")
	}

	log.Println("Fusing streams of gold disambiguation and ambiguous lattices")

	lFused := CombineLatticesStream(lDis, predAmbLatStream)

	log.Println()

	log.Println("Writing to output file", outMap)
	outFile, outFileError := os.Create(outMap)
	if outFileError != nil {
		panic(fmt.Sprintf("Couldn't create output file %s: %s", outMap, outFileError))
	}
	for fusedSent := range lFused {
		mdConfig := fusedSent.(*disambig.MDConfig)
		lat := mdConfig.Lattices
		output := lattice.Sentence2Lattice(lat, nil)
		tokens := make([]string, len(lat))
		for i, tok := range lat {
			tokens[i] = string(tok.Token)
		}
		lattice.UDWrite(outFile, []lattice.Lattice{output}, [][]string{tokens}, nil)
	}

	return nil
}

func FuseCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       Fuse,
		UsageLine: "fuse <file options> [arguments]",
		Short:     "Fuse morphological disambiguations into lattices",
		Long: `
Fuse morphological disambiguations into lattices

	$ ./yap fuse -d <disamb.> -l <amb. lat> -o <out output> [options]

`,
		Flag: *flag.NewFlagSet("md", flag.ExitOnError),
	}
	cmd.Flag.StringVar(&input, "l", "", "Lattices File (.conllul/.lattice)")
	cmd.Flag.StringVar(&inputGold, "d", "", "Morph. Disambigated File (.conll[u])")
	cmd.Flag.StringVar(&outMap, "o", "", "Output Lattice File")
	cmd.Flag.BoolVar(&FUSEuseConLLU, "conllu", true, "use CoNLL-U[L]-format")
	cmd.Flag.IntVar(&limit, "limit", 0, "limit input")
	return cmd
}
