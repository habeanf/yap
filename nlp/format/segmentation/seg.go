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
			if mapping.Token == NLP.ROOT_TOKEN {
				continue
			}
			writer.Write([]byte(mapping.Token))
			writer.Write([]byte{'\t'})
			morphForms := make([]string, len(mapping.Spellout))
			for i, morph := range mapping.Spellout {
				morphForms[i] = morph.Form
			}
			writer.Write([]byte(strings.Join(morphForms, ":")))
			writer.Write([]byte{'\n'})
		}
		writer.Write([]byte{'\n'})
	}
}

func WriteFile(filename string, graphs []NLP.MorphDependencyGraph) error {
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}
	Write(file, graphs)
	return nil
}
