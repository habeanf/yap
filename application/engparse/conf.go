package engparse

import (
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Conf struct {
	Values []string
}

func Read(reader io.Reader) (*Conf, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	lines = strings.Split(string(data), "\n")
	return &Conf{lines}
}

func ReadFile(filename string) (*Conf, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file)
}
