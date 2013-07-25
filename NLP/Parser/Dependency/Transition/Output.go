package Transition

import (
	"os"
	"text/tabwriter"
)

func (seq ConfigurationSequence) String() string {
	var buf bytes.Buffer
	w := new(tabwriter.Writer)
	w.Init(buf, 0, 8, 0, "\t", 0)
	seqLength := len(seq)
	for i, _ := range seq {
		conf := seq[seqLength-i-1]
		asString := seq.String()
		asBytes := []byte(asString)
		w.Write(asBytes)
	}
	w.Flush()
	return buf.String()
}
