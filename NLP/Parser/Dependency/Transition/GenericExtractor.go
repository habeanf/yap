package Transition

import (
	// . "chukuparser/Algorithm/Model"
	. "chukuparser/Algorithm/Model/Perceptron"
	. "chukuparser/Algorithm/Transition"
	// . "chukuparser/NLP"

	// "encoding/csv"
	// "os"
	"strings"
)

const (
	FEATURE_SEPARATOR   = ";"
	ATTRIBUTE_SEPARATOR = "."
)

type FeatureTemplateElement struct {
	Source     string
	Location   string
	Attributes []string

	ConfStr string

	srcAndLoc string
}

type FeatureTemplate []FeatureTemplateElement

type GenericExtractor struct {
	featureTemplates     []FeatureTemplate
	featureResultCache   map[string]string
	featureLocationCache map[string]*Attributes
}

// Verify GenericExtractor is a FeatureExtractor
var _ FeatureExtractor = GenericExtractor{}

func (x GenericExtractor) Features(instance *Instance) *[]Feature {
	conf, ok := instance.(Configuration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

	// Clear the feature element cache
	// the cache enables memoization of GetFeatureElement
	x.featureResultCache = make(map[string]string)
	x.featureLocationCache = make(map[string]*Attributes)

	features := make([]string, 0, x.EstimatedNumberOfFeatures())
	for i, template := range x.featureTemplates {
		feature, exists := x.GetFeature(conf, template)
		if exists {
			features = append(features, feature)
		}
	}
}

func (x GenericExtractor) EstimatedNumberOfFeatures() int {
	return len(x.featureTemplates)
}

func (x GenericExtractor) GetFeature(conf *Configuration, template FeatureTemplate) (string, bool) {
	featureValues := make([]string, 1, len(template)+1)
	featureValues[0] = template
	for _, templateElement := range template {
		// check if feature element was already computed
		cachedValue, cacheExists := x.featureResultCache[templateElement]
		if cacheExists {
			featureValues = append(featureValues, cachedValue)
		} else {
			elementValue, exists := x.GetFeatureElement(conf, templateElement)
			if !exists {
				return "", false
			}
			x.featureResultCache[templateElement] = elementValue
			featureValues = append(featureValues, elementValue)
		}
	}
	return strings.Join(featureValues, FEATURE_SEPARATOR), true
}

func (x GenericExtractor) GetFeatureElement(conf *Configuration, templateElement FeatureTemplateElement) (string, bool) {
	address, exists := conf.Address(byte[](templateElement.Address))
	if !exists {
		return "", false
	}
	attrValues, exists := make([]string, len(templateElement.Attributes))
	for i, attribute := range templateElement.Attributes {
		attrValue, exists = conf.Attribute(address, attribute)
		if !exists {
			return "", false
		}
		attrValues = append(attrValues, attValue)
	}
	return strings.Join(attrValues, ATTRIBUTE_SEPARATOR), true
}

func (x GenericExtractor) Load(filename string) {

}
