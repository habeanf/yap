package mapping

import (
	nlp "chukuparser/nlp/types"
	"fmt"
	"io"
	"os"
)

func Write(writer io.Writer, mappedSents []nlp.Mappings) {
	var curMorph int
	for _, mappedSent := range mappedSents {
		curMorph = 0
		for i, mapping := range mappedSent {
			if mapping.Token == nlp.ROOT_TOKEN {
				continue
			}
			for _, morph := range mapping.Spellout {
				writer.Write([]byte(fmt.Sprintf("%d\t%d\t", curMorph, curMorph+1)))
				writer.Write([]byte(morph.Form))
				writer.Write([]byte{'\t', '_', '\t'})
				writer.Write([]byte(morph.CPOS))
				writer.Write([]byte{'\t'})
				writer.Write([]byte(morph.POS))
				writer.Write([]byte{'\t'})
				if len(morph.FeatureStr) == 0 {
					writer.Write([]byte{'_'})
				} else {
					writer.Write([]byte(morph.FeatureStr))
				}
				writer.Write([]byte{'\t'})
				writer.Write([]byte(fmt.Sprintf("%d\n", i+1)))
				curMorph++
			}
		}
		writer.Write([]byte{'\n'})
	}
}

func WriteFile(filename string, mappedSents []nlp.Mappings) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, mappedSents)
	return nil
}
