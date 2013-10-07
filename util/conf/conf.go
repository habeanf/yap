package conf

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
	lines := strings.Split(string(data), "\n")
	retval := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 && line[0] != '#' {
			retval = append(retval, line)
		}
	}
	return &Conf{retval}, nil
}

func ReadFile(filename string) (*Conf, error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	return Read(file)
}
