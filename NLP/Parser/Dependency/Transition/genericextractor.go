package Transition

import (
	// . "chukuparser/Algorithm/Model"
	. "chukuparser/Algorithm/Model/Perceptron"

	// "encoding/csv"
	// "os"
	"strings"
)

const (
	FEATURE_SEPARATOR   = ";"
	ATTRIBUTE_SEPARATOR = "."
)

type FeatureTemplateElement struct {
	Address    string
	Attributes []string

	ConfStr string

	srcAndLoc string
}

type FeatureTemplate []FeatureTemplateElement

func (f FeatureTemplate) String() string {
	strs := make([]string, len(f))
	for i, featureElement := range f {
		strs[i] = featureElement.ConfStr
	}
	return strings.Join(strs, FEATURE_SEPARATOR)
}

type GenericExtractor struct {
	featureTemplates     []FeatureTemplate
	featureResultCache   map[string]string
	featureLocationCache map[string]Attributes
}

// Verify GenericExtractor is a FeatureExtractor
var _ FeatureExtractor = GenericExtractor{}

func (x GenericExtractor) Features(instance Instance) []Feature {
	conf, ok := instance.(DependencyConfiguration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

	// Clear the feature element cache
	// the cache enables memoization of GetFeatureElement
	x.featureResultCache = make(map[string]string)
	x.featureLocationCache = make(map[string]Attributes)

	features := make([]Feature, 0, x.EstimatedNumberOfFeatures())
	for _, template := range x.featureTemplates {
		feature, exists := x.GetFeature(conf, template)
		if exists {
			features = append(features, Feature(feature))
		}
	}
	return features
}

func (x GenericExtractor) EstimatedNumberOfFeatures() int {
	return len(x.featureTemplates)
}

func (x GenericExtractor) GetFeature(conf DependencyConfiguration, template FeatureTemplate) (string, bool) {
	featureValues := make([]string, 1, len(template)+1)
	featureValues[0] = template.String()
	for _, templateElement := range template {
		// check if feature element was already computed
		cachedValue, cacheExists := x.featureResultCache[templateElement.ConfStr]
		if cacheExists {
			featureValues = append(featureValues, cachedValue)
		} else {
			elementValue, exists := x.GetFeatureElement(conf, templateElement)
			if !exists {
				return "", false
			}
			x.featureResultCache[templateElement.ConfStr] = elementValue
			featureValues = append(featureValues, elementValue)
		}
	}
	return strings.Join(featureValues, FEATURE_SEPARATOR), true
}

func (x GenericExtractor) GetFeatureElement(conf DependencyConfiguration, templateElement FeatureTemplateElement) (string, bool) {
	address, exists := conf.Address([]byte(templateElement.Address))
	if !exists {
		return "", false
	}
	attrValues := make([]string, len(templateElement.Attributes))
	for i, attribute := range templateElement.Attributes {
		attrValue, exists := conf.Attribute(address, []byte(attribute))
		if !exists {
			return "", false
		}
		attrValues[i] = attrValue
	}
	return strings.Join(attrValues, ATTRIBUTE_SEPARATOR), true
}

func (x GenericExtractor) Load(filename string) {

}
