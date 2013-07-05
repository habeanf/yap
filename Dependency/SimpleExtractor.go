package Dependency

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
	source     string
	location   string
	properties []string
}

type FeatureTemplate []FeatureTemplateElement

type SimpleExtractor struct {
	featureTemplates []FeatureTemplate
}

func (x *SimpleExtractor) Features(instance Instance) *[]Feature {
	conf, ok := instance.(Configuration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

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

func (x *SimpleExtractor) GetFeature(conf Configuration, template FeatureTemplate) (string, bool) {
	featureValues := make([]string, len(template))
	for _, templateElement := range template {
		elementValue, exists := x.GetFeatureElement(conf, templateElement)
		if !exists {
			return "", false
		}
		featureValues = append(featureValues, elementValue)
	}
	return strings.Join(featureValues, FEATURE_SEPARATOR), true
}

func (x *SimpleExtractor) GetFeatureElement(conf Configuration, templateElement FeatureTemplateElement) (string, bool) {
	source := conf.GetSource(templateElement.source)
	if source == nil {
		return "", false
	}
	target, exists := conf.GetLocation(source, templateElement.location)
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

func (x *SimpleExtractor) Load(filename string) {

}
