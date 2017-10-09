package app

import (
	"yap/alg/search"
	"yap/alg/transition"
	transitionmodel "yap/alg/transition/model"
	"yap/nlp/format/conllu"
	"yap/nlp/format/lattice"
	"yap/nlp/format/lex"
	"yap/util"

	"yap/nlp/parser/disambig"
	"yap/nlp/parser/ma"
	nlp "yap/nlp/types"

	"fmt"
	"log"

	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
)

func HebPipeConfigOut() {
	log.Println("Configuration")
	log.Printf("Heb Lexicon:\t\t%s", prefixFile)
	log.Printf("Heb Prefix:\t\t%s", lexiconFile)
	log.Printf("OOV Strategy:\t%v", "Const:NNP")
	log.Println()
	if useConllU {
		log.Printf("CoNLL-U Input:\t%s", conlluFile)
	} else {
		log.Printf("Raw Input:\t\t%s", inRawFile)
	}
	log.Printf("Output:\t\t%s", outLatticeFile)
	log.Println()
}

func Mappings2TaggedStream(mappings chan interface{}, taggedSents chan interface{}, EMDWord, EMDPOS, EMDWPOS, EMDMHost, EMDMSuffix, EMDMorphProp, EMDTrans, EMDTokens *util.EnumSet) {
	var (
		i       int
		tagSent nlp.BasicETaggedSentence
		maps    nlp.Mappings
	)
	for instance := range mappings {
		maps = instance.(*disambig.MDConfig).Mappings
		tagSent = make([]nlp.EnumTaggedToken, 0, len(maps)*2)
		for _, curMap := range maps {
			for _, morph := range curMap.Spellout {
				taggedToken := nlp.EnumTaggedToken{
					TaggedToken: nlp.TaggedToken{
						Token: morph.Form,
						Lemma: morph.Lemma,
						POS:   morph.POS,
					},
					EToken:   morph.EForm,
					ELemma:   morph.ELemma,
					EPOS:     morph.EPOS,
					ETPOS:    morph.EFCPOS,
					EMHost:   morph.EMHost,
					EMSuffix: morph.EMSuffix,
				}
				tagSent = append(tagSent, taggedToken)
			}
		}
		taggedSents <- tagSent
		i++
	}
	close(taggedSents)
}

func HebPipe(cmd *commander.Command, args []string) error {
	useConllU = len(conlluFile) > 0
	var REQUIRED_FLAGS []string
	if useConllU {
		lattice.OVERRIDE_XPOS_WITH_UPOS = true
		REQUIRED_FLAGS = []string{"conllu", "out"}
	} else {
		REQUIRED_FLAGS = []string{"raw", "out"}
	}
	prefixLocation, found := util.LocateFile(prefixFile, DEFAULT_DATA_DIRS)
	if found {
		prefixFile = prefixLocation
	} else {
		REQUIRED_FLAGS = append(REQUIRED_FLAGS, "prefix")
	}
	lexiconLocation, found := util.LocateFile(lexiconFile, DEFAULT_DATA_DIRS)
	if found {
		lexiconFile = lexiconLocation
	} else {
		REQUIRED_FLAGS = append(REQUIRED_FLAGS, "lexicon")
	}
	VerifyFlags(cmd, REQUIRED_FLAGS)
	HebPipeConfigOut()
	// setup hebma
	if outFormat == "ud" {
		// override all skips in HEBLEX
		lex.SKIP_POLAR = false
		lex.SKIP_BINYAN = false
		lex.SKIP_ALL_TYPE = false
		lex.SKIP_TYPES = make(map[string]bool)
		lattice.IGNORE_LEMMA = false
		// Compatibility: No features for PROPN in UD Hebrew
		lex.STRIP_ALL_NNP_OF_FEATS = true
	}
	maData := new(ma.BGULex)
	maData.MAType = outFormat
	log.Println("Reading Morphological Analyzer BGU Prefixes")
	maData.LoadPrefixes(prefixFile)
	log.Println("Reading Morphological Analyzer BGU Lexicon")
	maData.LoadLex(lexiconFile, nnpnofeats)
	log.Println()

	// setup md
	var (
		mdTrans     transition.TransitionSystem
		mdModel     *transitionmodel.AvgMatrixSparse = &transitionmodel.AvgMatrixSparse{}
		modelExists bool
	)
	modelLocation, found := util.LocateFile(mdModelName, DEFAULT_MODEL_DIRS)
	if found {
		modelExists = true
		mdModelName = modelLocation
	} else {
		log.Println("Pre-trained model not found in default directories, looking for", mdModelName)
		modelExists = VerifyExists(mdModelName)
	}
	if !modelExists {
		log.Println("No model found")
		return nil
	}
	mdFeaturesLocation, found := util.LocateFile(mdFeaturesFile, DEFAULT_CONF_DIRS)
	if !found {
		log.Println("MD Features config not found")
		return nil
	}

	log.Println()
	log.Println("Found model file", mdModelName, " ... loading model")
	if useConllU {
		nlp.InitOpenParamFamily("UD")
	} else {
		nlp.InitOpenParamFamily("HEBTB")
	}
	serialization := ReadModel(mdModelName)
	mdModel.Deserialize(serialization.WeightModel)
	SetupMDEnum()
	EMDWord, EMDPOS, EMDWPOS, EMDMHost, EMDMSuffix, EMDMorphProp, EMDTrans, EMDTokens := serialization.EWord, serialization.EPOS, serialization.EWPOS, serialization.EMHost, serialization.EMSuffix, serialization.EMorphProp, serialization.ETrans, serialization.ETokens

	paramFunc, _ := nlp.MDParams[paramFuncName]
	mdTrans = &disambig.MDTrans{
		ParamFunc:   paramFunc,
		UsePOP:      true,
		POP:         POP,
		Transitions: ETrans,
	}

	mdTransitionSystem = transition.TransitionSystem(mdTrans)
	mdFeatureSetup, err := transition.LoadFeatureConfFile(mdFeaturesLocation)
	if err != nil {
		log.Println("Failed reading feature configuration file:", mdFeaturesFile)
		log.Fatalln(err)
	}
	mdFeatExtractor := SetupExtractor(mdFeatureSetup, []byte("MPL"))

	// setup configuration and beam
	mdConf := &disambig.MDConfig{
		ETokens:     EMDTokens,
		POP:         POP,
		Transitions: EMDTrans,
		ParamFunc:   paramFunc,
	}

	mdParser := &search.Beam{
		TransFunc:            mdTransitionSystem,
		FeatExtractor:        mdFeatExtractor,
		Base:                 mdConf,
		Size:                 mdBeamSize,
		ConcurrentExec:       ConcurrentBeam,
		Transitions:          EMDTrans,
		EstimatedTransitions: 1000, // chosen by random dice roll
		ShortTempAgenda:      true,
		Model:                mdModel,
	}

	// setup pipeline
	var (
		tokenStream chan nlp.BasicSentence
	)
	if useConllU {
		log.Println("Reading conllu from", conlluFile)
		inputStream, err := conllu.ReadFileAsStream(conlluFile, limit)
		if err != nil {
			panic(fmt.Sprintf("Failed reading CoNLL-U file - %v", err))
		}
		tokenStream = make(chan nlp.BasicSentence, 2)
		log.Println("Piping to conversion to []nlp.Token")
		go func() {
			var i int
			for sent := range inputStream {
				newSent := make([]nlp.Token, len(sent.Tokens))
				for j, token := range sent.Tokens {
					newSent[j] = nlp.Token(token)
				}
				i++
				tokenStream <- newSent
			}
			close(tokenStream)
		}()

	} else {
		return nil
	}
	stats := new(ma.AnalyzeStats)
	stats.Init()
	maData.Stats = stats
	maData.AlwaysNNP = alwaysnnp
	maData.LogOOV = showoov
	lattices := make(chan interface{}, 2)
	go func() {
		var i int
		for sent := range tokenStream {
			// log.SetPrefix(fmt.Sprintf("%v graph# %v ", prefix, i))
			lattice, _ := maData.Analyze(sent.Tokens())
			if i%100 == 0 {
				log.Println("At sent", i)
			}
			lattices <- lattice
			i++
		}
		close(lattices)
	}()
	// MA lattice -> MD
	mappings := make(chan interface{}, 2)
	go ParseStream(lattices, mappings, mdParser)

	// MD -> TaggedSentence
	taggedSents := make(chan interface{}, 2)
	go Mappings2TaggedStream(mappings, taggedSents, EMDWord, EMDPOS, EMDWPOS, EMDMHost, EMDMSuffix, EMDMorphProp, EMDTrans, EMDTokens)

	// Tagged -> Dep
	parsed := make(chan interface{}, 2)

	return nil
}

func HebPipeCmd() *commander.Command {
	cmd := &commander.Command{
		Run:       HebPipe,
		UsageLine: "hebpipe <file options> [arguments]",
		Short:     "run full hebrew pipeline: hebma -> md -> dep on raw input",
		Long: `
run lexicon-based morphological analyzer -> md -> dep on raw input

	$ ./yap hebpipe -conllu <conllu file> -out <output file> [options]

`,
		Flag: *flag.NewFlagSet("hebpipe", flag.ExitOnError),
	}
	// HEBMA config
	cmd.Flag.StringVar(&prefixFile, "prefix", "bgupreflex_withdef.utf8.hr", "Prefix file for morphological analyzer")
	cmd.Flag.StringVar(&lexiconFile, "lexicon", "bgulex.utf8.hr", "Lexicon file for morphological analyzer")
	cmd.Flag.StringVar(&inRawFile, "raw", "", "Input raw (tokenized) file")
	cmd.Flag.BoolVar(&alwaysnnp, "alwaysnnp", false, "Always add NNP to tokens and prefixed subtokens")
	cmd.Flag.BoolVar(&nnpnofeats, "addnnpnofeats", false, "Add NNP in lex but without features")
	cmd.Flag.BoolVar(&showoov, "showoov", false, "Output OOV tokens")
	cmd.Flag.BoolVar(&lex.LOG_FAILURES, "showlexerror", false, "Log errors encountered when loading the lexicon")

	// MD
	cmd.Flag.BoolVar(&ConcurrentBeam, "bconc", true, "Concurrent Beam")
	cmd.Flag.StringVar(&mdModelName, "mn", "hebmd.b32", "Modelfile")

	// DEP
	cmd.Flag.StringVar(&conlluFile, "conllu", "", "CoNLL-U-format input file")
	cmd.Flag.IntVar(&limit, "limit", 0, "Limit input set")
	cmd.Flag.StringVar(&outFormat, "format", "spmrl", "Output lattice format [spmrl|ud]")
	return cmd
}
