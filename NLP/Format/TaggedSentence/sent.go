package TaggedSentence

import (
	NLP "chukuparser/NLP/Types"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func Read(reader io.Reader) ([]NLP.TaggedSentence, error) {
	var (
		sent                            NLP.BasicTaggedSentence
		taggedTokenStrings, taggedToken []string
		token                           string
	)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	sentences := make([]NLP.TaggedSentence, len(lines)-1)
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		taggedTokenStrings = strings.Split(line, " ")
		if len(taggedTokenStrings) == 0 {
			return nil, errors.New("Empty sentence")
		}
		sent = make(NLP.BasicTaggedSentence, len(taggedTokenStrings))
		for j, taggedTokenString := range taggedTokenStrings {
			taggedToken = strings.Split(taggedTokenString, "/")
			if len(taggedToken) < 2 {
				return nil, errors.New("Got untagged token: " + taggedTokenString + " at line " + fmt.Sprintf("%v", i))
			}
			token = strings.Join(taggedToken[:len(taggedToken)-1], "/")
			sent[j] = NLP.TaggedToken{token, taggedToken[len(taggedToken)-1]}
		}
		sentences[i] = sent
	}
	return sentences, nil
}

func ReadFile(filename string) ([]NLP.TaggedSentence, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Read(file)
}
