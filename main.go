package main

import (
	"chukuparser/Util/Conll"
	"log"
)

func main() {
	s, _ := Conll.ReadFile("train5k.hebtb.sd.gold.conll")
	log.Println(len(s))
}
