package transition

import (
	"bufio"
	. "chukuparser/algorithm/featurevector"
	. "chukuparser/algorithm/perceptron"
	"chukuparser/util"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	// "sync"
)

const (
	FEATURE_SEPARATOR      = "+"
	ATTRIBUTE_SEPARATOR    = "|"
	TEMPLATE_PREFIX        = ":"
	GENERIC_SEPARATOR      = "|"
	REQUIREMENTS_SEPARATOR = ","
	APPROX_ELEMENTS        = 20
)

var (
	S0R2l, S0Rl int = -1, -1
)

type FeatureTemplateElement struct {
	Address    []byte
	Offset     int
	Attributes [][]byte

	ConfStr string
}

type FeatureTemplate struct {
	Elements                 []FeatureTemplateElement
	Requirements             []string
	ID                       int
	CachedElementIDs         []int // where to find the feature elements of the template in the cache
	CachedReqIDs             []int // cached address required to exist for element
	EWord, EPOS, EWPOS, ERel *util.EnumSet
}

func (f FeatureTemplate) String() string {
	strs := make([]string, len(f.Elements))
	for i, featureElement := range f.Elements {
		strs[i] = featureElement.ConfStr
	}
	return strings.Join(strs, FEATURE_SEPARATOR)
}

func (f FeatureTemplate) Format(value interface{}) string {
	if len(f.CachedElementIDs) == 1 {
		switch string(f.Elements[0].Attributes[0]) {
		case "w":
			return fmt.Sprintf("%v", f.EWord.ValueOf(value.(int)))
		case "p":
			return fmt.Sprintf("%v", f.EPOS.ValueOf(value.(int)))
		case "wp":
			return fmt.Sprintf("%v", f.EWPOS.ValueOf(value.(int)))
		case "l":
			return fmt.Sprintf("%d", value.(int)+1)
		default:
			return fmt.Sprint("%v", value)
		}
	} else {
		var sliceVal []interface{}
		switch valueType := value.(type) {
		case [2]interface{}:
			sliceVal = valueType[0:len(valueType)]
		case [3]interface{}:
			sliceVal = valueType[0:len(valueType)]
		case [4]interface{}:
			sliceVal = valueType[0:len(valueType)]
		case [5]interface{}:
			sliceVal = valueType[0:len(valueType)]
		default:
			panic("Don't know what to do")
		}
		retval := make([]string, len(f.CachedElementIDs))
		var attribNum int
		for _, element := range f.Elements {
			for _, attrib := range element.Attributes {
				value := sliceVal[attribNum]
				switch string(attrib) {
				case "w":
					if value == nil {
						value = 0
					}
					retval[attribNum] = fmt.Sprintf("%v", f.EWord.ValueOf(value.(int)))
				case "p":
					if value == nil {
						retval[attribNum] = "-NONE-"
					} else {
						retval[attribNum] = fmt.Sprintf("%v", f.EPOS.ValueOf(value.(int)))
					}
				case "wp":
					if value == nil {
						value = 0
					}
					ew := f.EWPOS.ValueOf(value.(int)).([2]string)
					retval[attribNum] = fmt.Sprintf("%s/%s", ew[0], ew[1])
				case "l":
					log.Println("Printing label")
					log.Println(value)
					if value == nil {
						retval[attribNum] = "-NONE-"
					} else {
						retval[attribNum] = fmt.Sprintf("%d", value.(int)+1)
					}
				case "d":
					if value != nil {
						retval[attribNum] = fmt.Sprintf("%d", value.(int))
					} else {
						retval[attribNum] = ""
					}
				case "vl", "vr":
					retval[attribNum] = fmt.Sprintf("%d", value.(int))
				case "sl", "sr":
					if value == nil {
						retval[attribNum] = "[ ]"
					}
					if value != nil {
						switch valType := value.(type) {
						case int:
							retval[attribNum] = fmt.Sprintf("[ %v ]", f.ERel.ValueOf(valType))
						case []int:
							set := valType
							tags := make([]string, len(set))
							for i, tag := range set {
								tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
							}
							retval[attribNum] = fmt.Sprintf("[ %s ]", strings.Join(tags, " "))
						case [2]int:
							set := valType[:]
							tags := make([]string, len(set))
							for i, tag := range set {
								tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
							}
							retval[attribNum] = fmt.Sprintf("[ %s ]", strings.Join(tags, " "))
						case [3]int:
							set := valType[:]
							tags := make([]string, len(set))
							for i, tag := range set {
								tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
							}
							retval[attribNum] = fmt.Sprintf("[ %s ]", strings.Join(tags, " "))
						case [4]int:
							set := valType[:]
							tags := make([]string, len(set))
							for i, tag := range set {
								tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
							}
							retval[attribNum] = fmt.Sprintf("[ %s ]", strings.Join(tags, " "))
						case [5]int:
							set := valType[:]
							tags := make([]string, len(set))
							for i, tag := range set {
								tags[i] = fmt.Sprintf("%v", f.ERel.ValueOf(tag))
							}
							retval[attribNum] = fmt.Sprintf("[ %s ]", strings.Join(tags, " "))

						default:
							panic("Don't know what to do with label set")
						}
					}
				default:
					panic("Don't know what to do with attribute")
					retval[attribNum] = fmt.Sprint("%v", value)
				}
				attribNum++
			}
		}
		return strings.Join(retval, " ")
	}
}

type GenericExtractor struct {
	FeatureTemplates []FeatureTemplate
	EFeatures        *util.EnumSet

	ElementEnum *util.EnumSet
	AddressEnum *util.EnumSet
	Elements    []FeatureTemplateElement

	Concurrent bool

	Log                      bool
	EWord, EPOS, EWPOS, ERel *util.EnumSet
}

// Verify GenericExtractor is a FeatureExtractor
var _ FeatureExtractor = &GenericExtractor{}

func (x *GenericExtractor) Init() {
	x.ElementEnum = util.NewEnumSet(APPROX_ELEMENTS)
	x.Elements = make([]FeatureTemplateElement, 0, APPROX_ELEMENTS)
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
	elementCache := make([]interface{}, len(x.Elements))
	attrArray := make([]interface{}, 0, 5)
	if S0R2l < 0 || S0Rl < 0 {
		panic("Did not set hard coded S0R2l or S0Rl")
	}
	// build element cache
	for i, elementTemplate := range x.Elements {
		element, exists := x.GetFeatureElement(conf, &elementTemplate, attrArray[0:0])
		if exists {
			if x.Log {
				log.Printf("%d %s: %v\n", i, elementTemplate.ConfStr, element)
			}
			if i == S0R2l { // un-documented code in zpar uses S0rl instead of S0r2l (wtf?!)
				elementCache[i] = elementCache[S0Rl]
			} else {
				elementCache[i] = element
			}
		} else {
			if x.Log {
				log.Printf("%d %s: nil\n", i, elementTemplate.ConfStr)
			}
			elementCache[i] = nil
		}
	}
	if x.Log {
		log.Println("Generating features:")
	}
	// generate features
	valuesArray := make([]interface{}, 0, 5)
	var (
		valuesSlice       []interface{}
		hasNilRequirement bool
	)
	for i, template := range x.FeatureTemplates {
		valuesSlice = valuesArray[0:0]
		hasNilRequirement = false
		if x.Log {
			log.Printf("Template %s; Requirements %v\n", template, template.Requirements)
		}
		for _, reqid := range template.CachedReqIDs {
			if elementCache[reqid] == nil {
				hasNilRequirement = true
				break
			}
		}
		if hasNilRequirement {
			features[i] = nil
		} else {
			for _, offset := range template.CachedElementIDs {
				if x.Log {
					log.Printf("\t(%d,%s): %v", offset, x.Elements[offset].ConfStr, elementCache[offset])
				}
				valuesSlice = append(valuesSlice, elementCache[offset])
			}
			val := GetArray(valuesSlice)
			features[i] = val
		}
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

func (x *GenericExtractor) ParseFeatureTemplate(featTemplateStr string, requirements string) (*FeatureTemplate, error) {
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
	reqArr := strings.Split(requirements, REQUIREMENTS_SEPARATOR)
	return &FeatureTemplate{Elements: featureTemplate, Requirements: reqArr,
		EWord: x.EWord, EPOS: x.EPOS, EWPOS: x.EWPOS, ERel: x.ERel}, nil
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
			*fullConfStr = string(element.Address) + "|" + string(attr)
			// log.Println("\t\tAttribute", *fullConfStr)
			elementId, isNew = x.ElementEnum.Add(*fullConfStr)
			if isNew {
				if *fullConfStr == "S0r2|l" {
					S0R2l = elementId
				}
				if *fullConfStr == "S0r|l" {
					S0Rl = elementId
				}
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
	feat.CachedReqIDs = make([]int, len(feat.Requirements))
	for i, req := range feat.Requirements {
		reqid, exists := x.ElementEnum.IndexOf(req)
		if !exists {
			panic(fmt.Sprintf("Can't find requirement element %s for features %s", req, feat))
		}
		feat.CachedReqIDs[i] = reqid
	}
}

func (x *GenericExtractor) LoadFeature(featTemplateStr string, requirements string) error {
	template, err := x.ParseFeatureTemplate(featTemplateStr, requirements)
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
		if err := x.LoadFeature(line, ""); err != nil {
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
