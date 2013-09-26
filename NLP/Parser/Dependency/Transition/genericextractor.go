package Transition

import (
	"bufio"
	. "chukuparser/Algorithm/FeatureVector"
	. "chukuparser/Algorithm/Perceptron"
	"chukuparser/Util"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	// "sync"
)

const (
	FEATURE_SEPARATOR   = "+"
	ATTRIBUTE_SEPARATOR = "|"
	TEMPLATE_PREFIX     = ":"
	GENERIC_SEPARATOR   = "|"
	APPROX_ELEMENTS     = 20
)

type FeatureTemplateElement struct {
	Address    []byte
	Offset     int
	Attributes [][]byte

	ConfStr string
}

type FeatureTemplate struct {
	Elements         []FeatureTemplateElement
	ID               int
	CachedElementIDs []int // where to find the feature elements of the template in the cache
	Requirement      int   // address required to exist for element
}

func (f FeatureTemplate) String() string {
	strs := make([]string, len(f.Elements))
	for i, featureElement := range f.Elements {
		strs[i] = featureElement.ConfStr
	}
	return strings.Join(strs, FEATURE_SEPARATOR)
}

type GenericExtractor struct {
	FeatureTemplates []FeatureTemplate
	EFeatures        *Util.EnumSet

	ElementEnum *Util.EnumSet
	AddressEnum *Util.EnumSet
	Elements    []FeatureTemplateElement

	ElementCache []interface{}

	Concurrent bool

	Log bool
}

// Verify GenericExtractor is a FeatureExtractor
var _ FeatureExtractor = &GenericExtractor{}

func (x *GenericExtractor) Init() {
	x.ElementEnum = Util.NewEnumSet(APPROX_ELEMENTS)
	x.Elements = make([]FeatureTemplateElement, 0, APPROX_ELEMENTS)
	x.ElementCache = make([]interface{}, 0, APPROX_ELEMENTS)
}

func (x *GenericExtractor) Features(instance Instance) []Feature {
	conf, ok := instance.(DependencyConfiguration)
	if !ok {
		panic("Type assertion that instance is a Configuration failed")
	}

	features := make([]Feature, len(x.FeatureTemplates))
	// if x.Concurrent {
	// 	featureChan := make(chan interface{})
	// 	wg := new(sync.WaitGroup)
	// 	for i, _ := range x.FeatureTemplates {
	// 		wg.Add(1)
	// 		go func(j int) {
	// 			defer wg.Done()
	// 			valuesArray := make([]interface{}, 0, 5)
	// 			attrArray := make([]interface{}, 0, 5)
	// 			featTemplate := x.FeatureTemplates[j]
	// 			feature, exists := x.GetFeature(conf, featTemplate, valuesArray, attrArray)
	// 			if exists {
	// 				featureChan <- feature
	// 			}
	// 		}(i)
	// 	}
	// 	go func() {
	// 		wg.Wait()
	// 		close(featureChan)
	// 	}()
	// 	for feature := range featureChan {
	// 		features = append(features, Feature(feature))
	// 	}
	// } else {
	if x.Log {
		log.Println("Generating elements:")
	}
	x.ElementCache = make([]interface{}, len(x.Elements))
	attrArray := make([]interface{}, 0, 5)
	// build element cache
	for i, elementTemplate := range x.Elements {
		element, exists := x.GetFeatureElement(conf, &elementTemplate, attrArray[0:0])
		if exists {
			if x.Log {
				log.Printf("%d %s: %v\n", i, elementTemplate.ConfStr, element)
			}
			x.ElementCache[i] = element
		} else {
			if x.Log {
				log.Printf("%d %s: nil\n", i, elementTemplate.ConfStr)
			}
			x.ElementCache[i] = nil
		}
	}
	if x.Log {
		log.Println("Generating features:")
	}
	// generate features
	valuesArray := make([]interface{}, 0, 5)
	var valuesSlice []interface{}
	for i, template := range x.FeatureTemplates {
		valuesSlice = valuesArray[0:0]
		if x.Log {
			log.Printf("Template %s\n", template)
		}
		for _, offset := range template.CachedElementIDs {
			if x.Log {
				log.Printf("\t(%d,%s): %v", offset, x.Elements[offset].ConfStr, x.ElementCache[offset])
			}
			valuesSlice = append(valuesSlice, x.ElementCache[offset])
		}
		val := GetArray(valuesSlice)
		features[i] = val
	}
	// valuesArray := make([]interface{}, 0, 5)
	// attrArray := make([]interface{}, 0, 5)
	// for _, tmpl := range x.FeatureTemplates {
	// 	feature, exists := x.GetFeature(conf, tmpl, valuesArray[0:0], attrArray[0:0])
	// 	if exists {
	// 		features = append(features, feature)
	// 	}
	// }
	// }
	return features
}

func (x *GenericExtractor) EstimatedNumberOfFeatures() int {
	return len(x.FeatureTemplates)
}

func (x *GenericExtractor) GetFeature(conf DependencyConfiguration, template FeatureTemplate, featureValues, attrValues []interface{}) (interface{}, bool) {
	// featureValues := make([]interface{}, 0, len(template.Elements))
	for i, templateElement := range template.Elements {
		featureValues = append(featureValues, nil)
		// check if feature element was already computed
		// cachedValue, cacheExists := x.FeatureResultCache[templateElement.ConfStr]
		cacheExists := false
		if cacheExists {
			// featureValues = append(featureValues, cachedValue)
		} else {
			attrValues = attrValues[0:0]
			elementValue, exists := x.GetFeatureElement(conf, &templateElement, attrValues[0:0])
			if !exists {
				return nil, false
			}
			// x.FeatureResultCache[templateElement.ConfStr] = elementValue
			featureValues[i] = elementValue
		}
	}
	if !x.Concurrent {
		return [3]interface{}{conf.Conf().GetLastTransition(), template.ID, GetArray(featureValues)}, true
	} else {
		return GetArray(featureValues), true
	}
}

func (x *GenericExtractor) GetFeatureElement(conf DependencyConfiguration, templateElement *FeatureTemplateElement, attrValues []interface{}) (interface{}, bool) {
	address, exists := conf.Address([]byte(templateElement.Address), templateElement.Offset)
	if !exists {
		return nil, false
	}
	// attrValues := make([]interface{}, len(templateElement.Attributes))
	for i, attribute := range templateElement.Attributes {
		attrValues = append(attrValues, nil)
		attrValue, exists := conf.Attribute(byte(templateElement.Address[0]), address, []byte(attribute))
		if !exists {
			return nil, false
		}
		attrValues[i] = attrValue
	}
	return GetArray(attrValues), true
}

func (x *GenericExtractor) ParseFeatureElement(featElementStr string) (*FeatureTemplateElement, error) {
	featElementStrPatchedWP := strings.Replace(featElementStr, "w|p", "wp", -1)
	elementParts := strings.Split(featElementStrPatchedWP, ATTRIBUTE_SEPARATOR)

	if len(elementParts) < 2 {
		return nil, errors.New("Not enough parts for element " + featElementStr)
	}

	// TODO: add validation to element parts
	element := new(FeatureTemplateElement)

	element.ConfStr = featElementStrPatchedWP
	element.Address = []byte(elementParts[0])
	// TODO fix to get more than one digit of offset
	parsedOffset, err := strconv.ParseInt(string(element.Address[1]), 10, 0)
	element.Offset = int(parsedOffset)
	if err != nil {
		panic("Error parsing feature element " + featElementStr + " " + err.Error())
	}
	element.Attributes = make([][]byte, len(elementParts)-1)

	for i, elementStr := range elementParts[1:] {
		element.Attributes[i] = []byte(elementStr)
	}
	return element, nil
}

func (x *GenericExtractor) ParseFeatureTemplate(featTemplateStr string) (*FeatureTemplate, error) {
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
	return &FeatureTemplate{Elements: featureTemplate}, nil
}

func (x *GenericExtractor) UpdateFeatureElementCache(feat *FeatureTemplate) {
	// log.Println("Update cache for", feat)
	feat.CachedElementIDs = make([]int, 0, len(feat.Elements))
	var (
		elementId int
		isNew     bool
	)
	for _, element := range feat.Elements {
		// log.Println("\tElement", element.ConfStr)
		for _, attr := range element.Attributes {
			fullConfStr := new(string)
			*fullConfStr = string(element.Address) + string(attr)
			// log.Println("\t\tAttribute", *fullConfStr)
			elementId, isNew = x.ElementEnum.Add(*fullConfStr)
			if isNew {
				fullElement := new(FeatureTemplateElement)
				fullElement.Address = element.Address
				fullElement.Offset = element.Offset
				fullElement.Attributes = make([][]byte, 1)
				fullElement.Attributes[0] = attr
				fullElement.ConfStr = *fullConfStr
				x.Elements = append(x.Elements, *fullElement)
				// log.Println("\t\tGenerated", fullElement.ConfStr)
			}
			// log.Println("\t\tID:", elementId)
			feat.CachedElementIDs = append(feat.CachedElementIDs, elementId)
		}
	}
}

func (x *GenericExtractor) LoadFeature(featTemplateStr string) error {
	template, err := x.ParseFeatureTemplate(featTemplateStr)
	if err != nil {
		return err
	}
	x.UpdateFeatureElementCache(template)
	template.ID, _ = x.EFeatures.Add(featTemplateStr)
	x.FeatureTemplates = append(x.FeatureTemplates, *template)
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

func GetArray(input []interface{}) interface{} {
	switch len(input) {
	case 0:
		return nil
	case 1:
		return input[0]
	case 2:
		return [2]interface{}{input[0], input[1]}
	case 3:
		return [3]interface{}{input[0], input[1], input[2]}
	case 4:
		return [4]interface{}{input[0], input[1], input[2], input[3]}
	case 5:
		return [5]interface{}{input[0], input[1], input[2], input[3], input[4]}
	case 6:
		return [6]interface{}{input[0], input[1], input[2], input[3], input[4], input[5]}
	default:
		result := make([]string, len(input))
		for i, val := range input {
			result[i] = fmt.Sprintf("%v", val)
		}
		return strings.Join(result, GENERIC_SEPARATOR)
	}
}

func GetArrayInt(input []int) interface{} {
	switch len(input) {
	case 0:
		return nil
	case 1:
		return input[0]
	case 2:
		return [2]int{input[0], input[1]}
	case 3:
		return [3]int{input[0], input[1], input[2]}
	case 4:
		return [4]int{input[0], input[1], input[2], input[3]}
	case 5:
		return [5]int{input[0], input[1], input[2], input[3], input[4]}
	case 6:
		return [6]int{input[0], input[1], input[2], input[3], input[4], input[5]}
	default:
		result := make([]string, len(input))
		for i, val := range input {
			result[i] = fmt.Sprintf("%v", val)
		}
		return strings.Join(result, GENERIC_SEPARATOR)
	}
}
