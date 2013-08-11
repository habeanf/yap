package Transition

import (
	"bufio"
	. "chukuparser/Algorithm/Model/Perceptron"
	"errors"
	"io"
	"strings"
	"sync"
)

const (
	FEATURE_SEPARATOR   = "+"
	ATTRIBUTE_SEPARATOR = "|"
	TEMPLATE_PREFIX     = ":"
)

type FeatureTemplateElement struct {
	Address    []byte
	Attributes [][]byte

	ConfStr string
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
	featureTemplates   []FeatureTemplate
	featureResultCache map[string]string
}

// Verify GenericExtractor is a FeatureExtractor
var _ FeatureExtractor = &GenericExtractor{}

func (x *GenericExtractor) Features(instance Instance) []Feature {
	conf, ok := instance.(DependencyConfiguration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

	// Clear the feature element cache
	// the cache enables memoization of GetFeatureElement
	x.featureResultCache = make(map[string]string)

	features := make([]Feature, 0, x.EstimatedNumberOfFeatures())

	featureChan := make(chan string)
	wg := new(sync.WaitGroup)
	for i, _ := range x.featureTemplates {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			featTemplate := x.featureTemplates[j]
			feature, exists := x.GetFeature(conf, featTemplate)
			if exists {
				featureChan <- feature
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(featureChan)
	}()
	for feature := range featureChan {
		features = append(features, Feature(feature))
	}
	return features
}

func (x *GenericExtractor) EstimatedNumberOfFeatures() int {
	return len(x.featureTemplates)
}

func (x *GenericExtractor) GetFeature(conf DependencyConfiguration, template FeatureTemplate) (string, bool) {
	featureValues := make([]string, 0, len(template))
	for _, templateElement := range template {
		// check if feature element was already computed
		// cachedValue, cacheExists := x.featureResultCache[templateElement.ConfStr]
		cacheExists := false
		if cacheExists {
			// featureValues = append(featureValues, cachedValue)
		} else {
			elementValue, exists := x.GetFeatureElement(conf, templateElement)
			if !exists {
				return "", false
			}
			// x.featureResultCache[templateElement.ConfStr] = elementValue
			featureValues = append(featureValues, elementValue)
		}
	}
	return template.String() + TEMPLATE_PREFIX + strings.Join(featureValues, FEATURE_SEPARATOR), true
}

func (x *GenericExtractor) GetFeatureElement(conf DependencyConfiguration, templateElement FeatureTemplateElement) (string, bool) {
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

func (x *GenericExtractor) ParseFeatureElement(featElementStr string) (*FeatureTemplateElement, error) {
	elementParts := strings.Split(featElementStr, ATTRIBUTE_SEPARATOR)

	if len(elementParts) < 2 {
		return nil, errors.New("Not enough parts for element " + featElementStr)
	}

	// TODO: add validation to element parts
	element := new(FeatureTemplateElement)

	element.ConfStr = featElementStr
	element.Address = []byte(elementParts[0])
	element.Attributes = make([][]byte, len(elementParts)-1)

	for i, elementStr := range elementParts[1:] {
		element.Attributes[i] = []byte(elementStr)
	}
	return element, nil
}

func (x *GenericExtractor) ParseFeatureTemplate(featTemplateStr string) (FeatureTemplate, error) {
	// remove any spaces
	featTemplateStr = strings.Replace(featTemplateStr, " ", "", -1)

	features := strings.Split(featTemplateStr, FEATURE_SEPARATOR)
	featureTemplate := make([]FeatureTemplateElement, len(features))

	for i, featElementStr := range features {
		parsedElement, err := x.ParseFeatureElement(featElementStr)
		if err != nil {
			return nil, err
		}
		featureTemplate[i] = *parsedElement
	}
	return FeatureTemplate(featureTemplate), nil
}

func (x *GenericExtractor) LoadFeature(featTemplateStr string) error {
	template, err := x.ParseFeatureTemplate(featTemplateStr)
	if err != nil {
		return err
	}
	x.featureTemplates = append(x.featureTemplates, template)
	return nil
}

func (x *GenericExtractor) LoadFeatures(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	// scan lines, lines beginning with # are ommitted
	for scanner.Scan() {
		line := scanner.Text()
		// skip blank and comment lines
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		// parse feature
		if err := x.LoadFeature(line); err != nil {
			return err
		}
	}
	return scanner.Err()
}
