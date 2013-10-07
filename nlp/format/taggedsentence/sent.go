package taggedsentence

import (
	nlp "chukuparser/nlp/types"
	"chukuparser/util"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func Read(reader io.Reader, EWord, EPOS, EWPOS *util.EnumSet) ([]nlp.EnumTaggedSentence, error) {
	var (
		sent                            nlp.BasicETaggedSentence
		taggedTokenStrings, taggedToken []string
		token, pos                      string
	)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	sentences := make([]nlp.EnumTaggedSentence, len(lines)-1)
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		taggedTokenStrings = strings.Split(line, " ")
		if len(taggedTokenStrings) == 0 {
			return nil, errors.New("Empty sentence")
		}
		sent = make(nlp.BasicETaggedSentence, len(taggedTokenStrings))
		for j, taggedTokenString := range taggedTokenStrings {
			taggedToken = strings.Split(taggedTokenString, "/")
			if len(taggedToken) < 2 {
				return nil, errors.New("Got untagged token: " + taggedTokenString + " at line " + fmt.Sprintf("%v", i))
			}
			token = strings.Join(taggedToken[:len(taggedToken)-1], "/")
			pos = taggedToken[len(taggedToken)-1]
			tokID, _ := EWord.Add(token)
			posID, _ := EPOS.Add(pos)
			tpID, _ := EWPOS.Add([2]string{token, pos})
			sent[j] = nlp.EnumTaggedToken{
				nlp.TaggedToken{token, pos},
				tokID, posID, tpID,
			}
		}
		sentences[i] = sent
	}
	return sentences, nil
}

func ReadFile(filename string, EWord, EPOS, EWPOS *util.EnumSet) ([]nlp.EnumTaggedSentence, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Read(file, EWord, EPOS, EWPOS)
}
