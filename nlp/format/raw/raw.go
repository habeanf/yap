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

func ReadStream(reader io.Reader, limit int) chan nlp.BasicSentence {
	sentences := make(chan nlp.BasicSentence, 2)

	go func() {
		var (
			i, numSentences int
		)
		bufReader := bufio.NewReader(reader)
		currentSent := make(nlp.BasicSentence, 0, 10)
		for curLine, isPrefix, err := bufReader.ReadLine(); err == nil; curLine, isPrefix, err = bufReader.ReadLine() {
			if isPrefix {
				panic("Buffer not large enough, fix me :(")
			}
			buf := bytes.NewBuffer(curLine)
			// log.Println("At record", i)
			// an empty line indicates a new record
			if len(curLine) == 0 {
				sentences <- currentSent
				numSentences++
				if limit > 0 && numSentences >= limit {
					close(sentences)
					return
				}
				currentSent = make(nlp.BasicSentence, 0, 10)
				continue
			} else {
				currentSent = append(currentSent, nlp.Token(buf.String()))
			}

			i++
		}
		close(sentences)
	}()
	return sentences
}

func Read(reader io.Reader, limit int) ([]nlp.BasicSentence, error) {
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
			if limit > 0 && len(sentences) >= limit {
				break
			}
			currentSent = make(nlp.BasicSentence, 0, 10)
		} else {
			currentSent = append(currentSent, nlp.Token(buf.String()))
		}

		i++
	}
	return sentences, nil
}

func ReadFile(filename string, limit int) ([]nlp.BasicSentence, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file, limit)
}

func Write(writer io.Writer, sents []interface{}) {
	for _, sent := range sents {
		for _, token := range sent.(nlp.BasicSentence) {
			writer.Write([]byte(token))
			writer.Write([]byte{'\n'})
		}
		writer.Write([]byte{'\n'})
	}
}
func WriteFile(filename string, sents []interface{}) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, sents)
	return nil
}

func ReadFileAsStream(filename string, limit int) (chan nlp.BasicSentence, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return ReadStream(file, limit), nil
}
