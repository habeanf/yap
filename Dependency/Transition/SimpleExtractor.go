package Transition

import (
	"encoding/csv"
	"os"
	"strings"
)

const (
	FEATURE_SEPARATOR  = ";"
	PROPERTY_SEPARATOR = "."
)

type FeatureTemplateElement struct {
	confStr    string
	source     string
	location   string
	srcAndLoc  string
	properties []string
}

type FeatureTemplate []FeatureTemplateElement

type SimpleExtractor struct {
	featureTemplates     []FeatureTemplate
	featureResultCache   map[string]string
	featureLocationCache map[string]*HasProperties
}

func (x *SimpleExtractor) Features(instance Instance) *[]Feature {
	conf, ok := instance.(Configuration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

	// Clear the feature element cache
	// the cache enables memoization of GetFeatureElement
	x.featureResultCache = make(map[string]string)
	x.featureLocationCache = make(map[string]*HasProperties)

	features := make([]string, x.NumberOfFeatures())
	for i, template := range x.featureTemplates {
		feature, exists := x.GetFeature(conf, template)
		if exists {
			features = append(features, feature)
		}
	}
}

func (x *SimpleExtractor) NumberOfFeatures() int {
	return len(x.featureTemplates)
}

func (x *SimpleExtractor) GetFeature(conf *Configuration, template FeatureTemplate) (string, bool) {
	featureValues := make([]string, len(template))
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

func (x *SimpleExtractor) GetFeatureElement(conf *Configuration, templateElement FeatureTemplateElement) (string, bool) {
	target, exists := x.GetLocation(conf, templateElement)
	if !exists {
		return "", false
	}
	propertyValues := make([]string, len(templateElement.properties))
	for i, property := range templateElement.properties {
		propertyValue, exists = target.GetProperty(property)
		if !exists {
			return "", false
		}
		propertyValues = append(propertyValues, propertyValue)
	}
	return strings.Join(propertyValues, PROPERTY_SEPARATOR), true
}

func (x *SimpleExtractor) GetLocation(conf *Configuration, templateElement FeatureTemplateElement) (*HasProperties, bool) {
	cachedLocation, exists := x.featureLocationCache[templateElement.srcAndLoc]
	if exists {
		return cachedLocation, true
	}
	source := conf.GetSource(templateElement.source)
	if source == nil {
		return nil, false
	}
	target, exists := conf.GetLocation(source, []byte(templateElement.location))
	if !exists {
		return nil, false
	}
	x.featureLocationCache[templateElement.srcAndLoc] = target
	return target, true
}

func (x *SimpleExtractor) Load(filename string) {

}
