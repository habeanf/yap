package mapping

import (
	"chukuparser/nlp/parser/disambig"
	nlp "chukuparser/nlp/types"
	"fmt"
	"io"
	"os"

	// "log"
)

func WriteMorph(writer io.Writer, morph *nlp.EMorpheme, curMorph, curToken int) {
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
	writer.Write([]byte(fmt.Sprintf("%d\n", curToken+1)))
}

func Write(writer io.Writer, mappedSents []interface{}) {
	var curMorph int
	for _, mappedSent := range mappedSents {
		curMorph = 0
		for i, mapping := range mappedSent.(*disambig.MDConfig).Mappings {
			// log.Println("At token", i, mapping.Token)
			if mapping.Token == nlp.ROOT_TOKEN {
				continue
			}
			// if mapping.Spellout != nil {
			// 	log.Println("\t", mapping.Spellout.AsString())
			// } else {
			// 	log.Println("\t", "*No spellout")
			// }
			for _, morph := range mapping.Spellout {
				if morph == nil {
					// log.Println("\t", "Morph is nil, continuing")
					continue
				}
				if morph.Concat {
					WriteMorph(writer, morph.OrigMorph, curMorph, i)
					curMorph++
					WriteMorph(writer, morph.SwallowedMorph, curMorph, i)
				} else {
					WriteMorph(writer, morph, curMorph, i)
				}
				// log.Println("\t", "At morph", j, morph.Form)
				curMorph++
			}
		}
		writer.Write([]byte{'\n'})
	}
}

func WriteFile(filename string, mappedSents []interface{}) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, mappedSents)
	return nil
}
