package raw

// Package raw reads raw format files
// raw files contain a token per line
// sentences end with a new line

import (
	nlp "yap/nlp/types"

	"bufio"
	"bytes"
	"io"
	// "log"
	"os"
)

func Read(reader io.Reader) ([]nlp.BasicSentence, error) {
	var sentences []nlp.BasicSentence
	bufReader := bufio.NewReader(reader)

	var (
		i int
	)
	currentSent := make(nlp.BasicSentence, 0, 10)
	for curLine, isPrefix, err := bufReader.ReadLine(); err == nil; curLine, isPrefix, err = bufReader.ReadLine() {
		if isPrefix {
			panic("Buffer not large enough, fix me :(")
		}
		buf := bytes.NewBuffer(curLine)
		// log.Println("At record", i)
		// an empty line indicates a new record
		if len(curLine) == 0 {
			sentences = append(sentences, currentSent)
			currentSent = make(nlp.BasicSentence, 0, 10)
		} else {
			currentSent = append(currentSent, nlp.Token(buf.String()))
		}

		i++
	}
	return sentences, nil
}

func ReadFile(filename string) ([]nlp.BasicSentence, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file)
}
