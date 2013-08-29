package Application

// func oldmain() {
// 	trainFile, trainSeqFile := "train.conll", "train.gob"
// 	inputFile, outputFile := "devi.txt", "devo.txt"
// 	iterations, beamSize := 1, 64

// 	modelFile := fmt.Sprintf("model.b%d.i%d", beamSize, iterations)

// 	var goldSequences []Perceptron.DecodedInstance

// 	RegisterTypes()
// 	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
// 	log.Println("Configuration")
// 	log.Println("Train file:\t", trainFile)
// 	log.Println("Seq file:\t", trainSeqFile)
// 	log.Println("Input file:\t", inputFile)
// 	log.Println("Output file:\t", outputFile)
// 	log.Println("Iterations:\t", iterations)
// 	log.Println("Beam Size:\t", beamSize)
// 	log.Println("Model file:\t", modelFile)
// 	runtime.GOMAXPROCS(runtime.NumCPU())
// 	// runtime.GOMAXPROCS(1)

// 	// launch net server for profiling
// 	go func() {
// 		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
// 	}()

// 	log.Println("Reading training sentences from", trainFile)
// 	s, e := Conll.ReadFile(trainFile)
// 	if e != nil {
// 		log.Println(e)
// 		return
// 	}
// 	log.Println("Read", len(s), "sentences from", trainFile)
// 	log.Println("Converting from conll to internal format")
// 	// goldGraphs := Conll.Conll2GraphCorpus(s)
// 	var goldGraphs []*Morph.BasicMorphGraph
// 	log.Println("Parsing with gold to get training sequences")
// 	goldSequences = TrainingSequences(goldGraphs, RICH_FEATURES)
// 	// log.Println("Writing training sequences to", trainSeqFile)
// 	// WriteTraining(goldSequences, trainSeqFile)
// 	// log.Println("Loading training sequences from", trainSeqFile)
// 	// goldSequences = ReadTraining(trainSeqFile)
// 	log.Println("Successfully Loaded", len(goldSequences), "training sequences")
// 	log.Println("Running GC")
// 	log.Println("Before")
// 	Util.LogMemory()
// 	runtime.GC()
// 	log.Println("After")
// 	Util.LogMemory()
// 	log.Println("Training", iterations, "iteration(s)")
// 	model := Train(goldSequences, iterations, beamSize, RICH_FEATURES, modelFile)
// 	log.Println("Done Training")
// 	Util.LogMemory()

// 	log.Println("Writing final model to", modelFile)
// 	WriteModel(model, modelFile)
// 	// model := ReadModel(modelFile)
// 	// log.Println("Read model from", modelFile)
// 	// sents, e2 := TaggedSentence.ReadFile(inputFile)
// 	// log.Println("Read", len(sents), "from", inputFile)
// 	// if e2 != nil {
// 	// 	log.Println(e2)
// 	// 	return
// 	// }
// 	var sents []NLP.LatticeSentence
// 	log.Print("Parsing")
// 	parsedGraphs := Parse(sents, beamSize, Dependency.ParameterModel(&PerceptronModel{model}), RICH_FEATURES)
// 	log.Println("Converting to conll")
// 	// graphAsConll := Conll.Graph2ConllCorpus(parsedGraphs)
// 	log.Println("Wrote", len(parsedGraphs), "in conll format to", outputFile)
// 	// Conll.WriteFile(outputFile, graphAsConll)
// }
