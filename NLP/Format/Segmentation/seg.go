package Segmentation

import (
	NLP "chukuparser/NLP/Types"
	"io"
	"os"
	"strings"
)

func Write(writer io.Writer, graphs []NLP.MorphDependencyGraph) {
	for _, graph := range graphs {
		for _, mapping := range graph.GetMappings() {
			writer.Write(mapping.Token)
			writer.Write("\t")
			morphForms := make([]string, len(mapping.Spellout))
			for i, morph := range mapping.Spellout {
				morphForms[i] = morph.Form
			}
			writer.Write(strings.Join(morphForms, ":"))
			writer.Write("\n")
		}
		writer.Write([]byte{'\n'})
	}
}

func WriteFile(filename string, graph []NLP.MorphDependencyGraph) {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, sents)
	return nil
}
