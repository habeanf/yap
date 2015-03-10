package disambig

import (
	// "chukuparser/alg/graph"
	. "chukuparser/alg"
	. "chukuparser/alg/transition"
	nlp "chukuparser/nlp/types"
	"chukuparser/util"

	"fmt"
	"log"
	// "reflect"
	"sort"
	"strings"
)

type MDConfig struct {
	LatticeQueue Queue
	Lattices     nlp.LatticeSentence
	Mappings     nlp.Mappings
	Morphemes    nlp.Morphemes

	CurrentLatNode int

	InternalPrevious Configuration
	Last             Transition
	ETokens          *util.EnumSet
	Log              bool

	POP    Transition
	popped int
}

var _ Configuration = &MDConfig{}

func (c *MDConfig) Init(abstractLattice interface{}) {
	latticeSent := abstractLattice.(nlp.LatticeSentence)
	sentLength := len(latticeSent)

	c.Lattices = latticeSent

	maxSentLength := 0
	var latP *nlp.Lattice
	for _, lat := range c.Lattices {
		latP = &lat
		maxSentLength += latP.MaxPathLen()
	}

	c.LatticeQueue = NewQueueSlice(sentLength)

	// push indexes of statement nodes to *LatticeQueue*, in reverse order (first word at the top of the queue)
	for i := 0; i < sentLength; i++ {
		c.LatticeQueue.Enqueue(i)
	}

	// initialize first mapping structure
	c.Mappings = make([]*nlp.Mapping, 1, len(c.Lattices))
	c.Mappings[0] = &nlp.Mapping{c.Lattices[0].Token, make(nlp.Spellout, 0, 1)}
	c.Morphemes = make(nlp.Morphemes, 0, len(c.Lattices)*2)
	// explicit resetting of zero-valued properties
	// in case of reuse
	c.Last = 0
	c.popped = 0
}

func (c *MDConfig) Terminal() bool {
	// return c.Last == Transition(0) && c.Alignment() == 1
	return c.LatticeQueue.Size() == 0 && c.popped == len(c.Mappings)
}

func (c *MDConfig) Copy() Configuration {
	newConf := new(MDConfig)
	c.CopyTo(newConf)
	return newConf
}

func (c *MDConfig) CopyTo(target Configuration) {
	newConf, ok := target.(*MDConfig)
	if !ok {
		panic("Can't copy into non *MDConfig")
	}
	newConf.ETokens = c.ETokens
	newConf.Mappings = make([]*nlp.Mapping, len(c.Mappings), util.Max(cap(c.Mappings), len(c.Lattices)))
	copy(newConf.Mappings, c.Mappings)

	newConf.Morphemes = make(nlp.Morphemes, len(c.Morphemes), cap(c.Morphemes))
	copy(newConf.Morphemes, c.Morphemes)

	// verify initialization (base configurations are not initialized)
	if len(c.Mappings) > 0 {
		// also copy a new spellout of the current mapping
		lastMappingIdx := len(c.Mappings) - 1
		newLastMapping := &nlp.Mapping{c.Mappings[lastMappingIdx].Token,
			make(nlp.Spellout, len(c.Mappings[lastMappingIdx].Spellout), cap(c.Mappings[lastMappingIdx].Spellout))}
		copy(newLastMapping.Spellout, c.Mappings[lastMappingIdx].Spellout)
		newConf.Mappings[lastMappingIdx] = newLastMapping
	}

	if c.LatticeQueue != nil {
		newConf.LatticeQueue = c.LatticeQueue.Copy()
	}
	// lattices slice is read only, no need for copy
	newConf.Lattices = c.Lattices
	newConf.InternalPrevious = c
	newConf.CurrentLatNode = c.CurrentLatNode
	newConf.popped = c.popped
	newConf.POP = c.POP
}

func (c *MDConfig) GetSequence() ConfigurationSequence {
	if c.Mappings == nil {
		return make(ConfigurationSequence, 0)
	}
	retval := make(ConfigurationSequence, 0, len(c.Mappings))
	currentConf := c
	for {
		retval = append(retval, currentConf)
		if currentConf.InternalPrevious == nil {
			break
		} else {
			currentConf = currentConf.InternalPrevious.(*MDConfig)
		}
	}
	return retval
}

func (c *MDConfig) SetLastTransition(t Transition) {
	c.Last = t
}

func (c *MDConfig) GetLastTransition() Transition {
	return Transition(c.Last)
}

func (c *MDConfig) String() string {
	if c.Mappings == nil {
		return fmt.Sprintf("\t=>([],\t[]) - %v", c.Alignment())
	}
	mapLen := len(c.Mappings)
	transStr := "MD"
	if c.Last == Transition(0) {
		transStr = "IDLE"
	}
	if c.Last == c.POP {
		transStr = "POP"
	}
	if mapLen > 0 || len(c.Mappings[mapLen-1].Spellout) > 0 {
		if mapLen == 1 && len(c.Mappings[mapLen-1].Spellout) == 0 {
			return fmt.Sprintf("\t=>([%s],\t[]) - %v", c.StringLatticeQueue(), c.Alignment())
		}
		if len(c.Mappings[mapLen-1].Spellout) > 0 && len(c.Mappings[mapLen-1].Spellout) > 0 {
			return fmt.Sprintf("%s\t=>([%s],\t[%v]) - %v", transStr, c.StringLatticeQueue(), c.Mappings[mapLen-1], c.Alignment())
		}
		return fmt.Sprintf("%s\t=>([%s],\t[%v]) - %v", transStr, c.StringLatticeQueue(), c.Mappings[mapLen-2], c.Alignment())
	} else {
		return fmt.Sprintf("\t=>([%s],\t[%s]) - %v", c.StringLatticeQueue(), "", c.Alignment())
	}
}

func (c *MDConfig) StringLatticeQueue() string {
	if c.LatticeQueue == nil {
		return ""
	}
	queueSize := c.LatticeQueue.Size()
	switch {
	case queueSize > 0 && queueSize <= 3:
		var queueStrings []string = make([]string, 0, 3)
		for i := 0; i < c.LatticeQueue.Size(); i++ {
			atI, _ := c.LatticeQueue.Index(i)
			queueStrings = append(queueStrings, fmt.Sprintf("%v - %v", string(c.Lattices[atI].Token), c.CurrentLatNode))
		}
		return strings.Join(queueStrings, ",")
	case queueSize > 3:
		headID, _ := c.LatticeQueue.Index(0)
		tailID, _ := c.LatticeQueue.Index(c.LatticeQueue.Size() - 1)
		head := c.Lattices[headID]
		tail := c.Lattices[tailID]
		return strings.Join([]string{string(head.Token), "...", string(tail.Token)}, ",")
	default:
		return ""
	}

}
func (c *MDConfig) Equal(otherEq util.Equaler) bool {
	if (otherEq == nil && c != nil) || (c == nil && otherEq != nil) {
		if c.Log {
			log.Println("\tfalse default")
		}
		return false
	}
	// c.Log = true
	switch other := otherEq.(type) {
	case *MDConfig:
		if (other == nil && c != nil) || (c == nil && other != nil) {
			if c.Log {
				log.Println("\tfalse 0")
			}
			return false
		}
		if c.Log {
			log.Println("Comparing", c, "to", other)
			log.Println("Comparing\n", c.GetSequence(), "\n\tto\n", other.GetSequence())
		}
		if other.Last != c.Last {
			if c.Log {
				log.Println("\tfalse 1")
			}
			return false
		}
		if c.InternalPrevious == nil && other.InternalPrevious == nil {
			if c.Log {
				log.Println("\ttrue")
			}
			return true
		}
		if c.InternalPrevious != nil && other.InternalPrevious != nil {
			if c.Log {
				log.Println("\trecurse")
				c.InternalPrevious.(*MDConfig).Log = c.Log
			}
			return c.InternalPrevious.Equal(other.InternalPrevious)
		} else {
			if c.Log {
				log.Println("\tfalse 3: ", c.InternalPrevious, "vs", other.InternalPrevious)
			}
			return false
		}
	default:
		panic("TODO: Figure out what the type of the other is ([]*nlp.Mapping?)")
	}
}

func (c *MDConfig) Previous() Configuration {
	return c.InternalPrevious
}

func (c *MDConfig) SetPrevious(prev Configuration) {
	c.InternalPrevious = prev
}

func (c *MDConfig) Clear() {
	c.InternalPrevious = nil
}

func (c *MDConfig) AddMapping(m *nlp.EMorpheme) {
	// log.Println("\tAdding mapping to spellout")
	c.CurrentLatNode = m.To()

	currentLatIdx, _ := c.LatticeQueue.Peek()

	if len(c.Mappings) < currentLatIdx {
		// log.Println("\tAdding new mapping because", len(c.Mappings), currentLatIdx)
	}

	currentMap := c.Mappings[len(c.Mappings)-1]
	currentMap.Spellout = append(currentMap.Spellout, m)

	// log.Println("\tNode bumped to", c.CurrentLatNode, "of lattice", currentLatIdx)
	// debugLat := c.Lattices[currentLatIdx]
	// log.Println("\tCurrent lattice token bottom/top", debugLat.Token, debugLat.Bottom(), debugLat.Top())
	// if current lattice node is the last of current lattice
	// then pop lattice and make new mapping struct
	if currentLat := c.Lattices[currentLatIdx]; c.CurrentLatNode == currentLat.Top() {
		// log.Println("\tPopping lattice queue")
		c.LatticeQueue.Pop()
		val, exists := c.LatticeQueue.Peek()
		// log.Println("\tNow at lattice (exists)", val, exists)
		if exists {
			c.Mappings = append(c.Mappings, &nlp.Mapping{c.Lattices[val].Token, make(nlp.Spellout, 0, 1)})
			// log.Println("\tSetting token to", c.Lattices[val].Token)
		}
	}
	c.Morphemes = append(c.Morphemes, m)
}

func (c *MDConfig) Address(location []byte, sourceOffset int) (int, bool, bool) {
	source := c.GetSource(location[0])
	if source == nil {
		return 0, false, false
	}
	var (
		atAddress int
		exists    bool
	)
	// test if feature address is a generator of feature (e.g. for each child..)
	locationLen := len(location)
	if location[0] == 'L' && locationLen >= 4 {
		if string(location[2:4]) == "Ci" {
			return atAddress, true, true
		}
	}
	sourceOffsetInt := int(sourceOffset)
	// hack for lattices to retrieve previously seen lattices
	if location[0] == 'L' && sourceOffsetInt < 0 {
		// assumes lattice indices are continuous in lattice queue
		atAddress, exists = source.Index(0)
		atAddress = atAddress + sourceOffsetInt
		if exists {
			exists = atAddress >= 0
		}
	} else {
		atAddress, exists = source.Index(sourceOffsetInt)
	}
	if !exists {
		return 0, false, false
	}

	location = location[2:]
	if len(location) == 0 {
		return atAddress, true, false
	}
	return 0, false, false
}

func (c *MDConfig) Attribute(source byte, nodeID int, attribute []byte) (interface{}, bool) {
	switch source {
	case 'M':
		morpheme := c.Morphemes[nodeID]
		switch attribute[0] {
		case 'm':
			if len(attribute) > 1 && attribute[1] == 'p' {
				return morpheme.EFCPOS, true
			} else {
				return morpheme.EForm, true
			}
		case 'p':
			return morpheme.EPOS, true
		case 'f':
			return morpheme.EFeatures, true
		case 't':
			lat := c.Lattices[morpheme.TokenID]
			tokId, _ := c.ETokens.Add(lat.Token)
			return tokId, true
		case 'i': // path of lattice of last morpheme
			result := make([]string, 0, 5) // assume most lattice lengths are <= 5
			// log.Println("Generating idle feature starting with morpheme")
			// log.Println(" mappings are")
			// log.Println(c.Mappings)
			// log.Println(" morphemes are (current nodeID is:", nodeID, ")")
			// log.Println(c.Morphemes)
			// log.Println(morpheme)
			curTokenId := morpheme.TokenID
			// log.Println("Token is", curTokenId, c.Lattices[curTokenId-1].Token)
			for {
				// log.Println("Adding morph string", nlp.Funcs_Main_POS_Both_Prop(morpheme))
				result = append(result, nlp.Funcs_Main_POS_Both_Prop(morpheme))
				// get the next morpheme
				// break if reached end of morpheme stack or reached
				// next token (== lattice)
				nodeID--
				if nodeID < 0 {
					break
				}
				morpheme = c.Morphemes[nodeID]
				if morpheme.TokenID != curTokenId {
					break
				}
			}
			// log.Println("Idle feature", fmt.Sprintf("%v", result))
			return fmt.Sprintf("%v", result), true
		}
	case 'L':
		lat := c.Lattices[nodeID]
		switch attribute[0] {
		case 't': // token of last lattice
			tokId, _ := c.ETokens.Add(lat.Token)
			return tokId, true
		case 'n': // next edges of current lattice node
			if nextEdges, exists := lat.Next[c.CurrentLatNode]; exists {
				retval := make([]string, len(nextEdges))
				for _, edgeId := range nextEdges {
					curEdge := lat.Morphemes[edgeId]
					retval = append(retval, nlp.Funcs_Main_POS_Both_Prop(curEdge))
				}
				sort.StringSlice(retval).Sort()
				return fmt.Sprintf("%v", retval), true
			}
		case 'i': // path of lattice
			// log.Println("Generating idle feature starting with morpheme")
			// log.Println(" mappings are")
			// log.Println(c.Mappings)
			// log.Println(" morphemes are (current nodeID is:", nodeID, ")")
			// log.Println(c.Morphemes)
			latMapping := c.Mappings[nodeID]
			result := make([]string, len(latMapping.Spellout)) // assume most lattice lengths are <= 5
			for i, morpheme := range latMapping.Spellout {
				// log.Println("Adding morph string", nlp.Funcs_Main_POS_Both_Prop(morpheme))
				result[i] = nlp.Funcs_Main_POS_Both_Prop(morpheme)
				// get the next morpheme
				// break if reached end of morpheme stack or reached
				// next token (== lattice)
			}
			return fmt.Sprintf("%v-%v", result, lat.Token), true
		}
	}
	return 0, false
}

func (c *MDConfig) GenerateAddresses(nodeID int, location []byte) (nodeIDs []int) {
	return util.RangeInt(len(c.Lattices))
}

func (c *MDConfig) GetSource(location byte) Index {
	switch location {
	case 'M':
		return c.Morphemes
	case 'L':
		return c.LatticeQueue
	}
	return nil
}

func (c *MDConfig) Alignment() int {
	// return c.popped
	if c.popped == len(c.Mappings) && c.LatticeQueue.Size() > 0 {
		return 0
	} else {
		return 1
	}
	// return len(c.Mappings)
}

func (c *MDConfig) Assignment() uint16 {
	return uint16(len(c.Mappings))
}

func (c *MDConfig) Len() int {
	if c == nil {
		return 0
	}
	if c.Previous() != nil {
		return 1 + c.Previous().Len()
	} else {
		return 1
	}
}

func (c *MDConfig) Pop() {
	c.popped += 1
}
