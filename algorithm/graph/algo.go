package graph

import (
	"fmt"
	"log"
)

func YieldAllPaths(l DirectedGraph, from, to int) chan []DirectedEdge {
	pathChan := make(chan []DirectedEdge)
	go func() {
		// preempt a graph with only one edge
		if l.NumberOfEdges() == 0 {
			close(pathChan)
			return
		}
		if l.NumberOfEdges() == 1 {
			pathChan <- []DirectedEdge{l.GetDirectedEdge(0)}
			close(pathChan)
			return
		}
		// lookup outgoing edge IDs by id
		outgoing := make(map[int][]DirectedEdge, to-from)
		for _, edgeId := range l.GetEdges() {
			edge := l.GetDirectedEdge(edgeId)
			outSet, exists := outgoing[edge.From()]
			if !exists {
				outSet = make([]DirectedEdge, 0, to-from)
			}
			outSet = append(outSet, edge)
			outgoing[edge.From()] = outSet
		}
		var (
			// TODO: better heuristic for agenda capacity from data sampling
			// histogram "width" as a function of max length
			agenda                 [][]DirectedEdge = make([][]DirectedEdge, 0, to-from)
			cur, copyList, listOut []DirectedEdge
			lastEdge, next         DirectedEdge
			exists                 bool
		)
		for _, initialEdge := range outgoing[from] {
			initialPath := make([]DirectedEdge, 1, to-from)
			initialPath[0] = initialEdge
			if initialEdge.To() == to {
				pathChan <- initialPath
			} else {
				agenda = append(agenda, initialPath)
			}
		}
		var num int
		for len(agenda) > 0 {
			// log.Println("Agenda is:", agenda)
			// pop agenda
			cur = agenda[len(agenda)-1]
			agenda = agenda[:len(agenda)-1]

			// get last
			lastEdge = cur[len(cur)-1]
			// log.Println("Getting outgoing for", lastEdge)
			listOut, exists = outgoing[lastEdge.To()]
			// log.Println("Outgoing for", lastEdge, "is", listOut)
			if !exists {
				listOut, exists = outgoing[lastEdge.To()+1]
			}
			if exists {
				for _, next = range listOut {
					copyList = make([]DirectedEdge, len(cur)+1)
					copy(copyList, cur)
					copyList[len(cur)] = next
					// log.Println("\tGot new path", copyList)
					if next.To() == to {
						// log.Println("\tYielding", copyList)
						pathChan <- copyList
					} else {
						// log.Println("\tAgenda Push", copyList)
						agenda = append(agenda, copyList)
					}
				}
			} else {
				panic(fmt.Sprintf("Graph is not a lattice, got missing outgoing node %v", lastEdge.To()))
			}
			num++
			if num > 500 {
				log.Println("Breaking on over 500 agenda iterations")
				break
			}
		}
		close(pathChan)
	}()
	return pathChan
}
