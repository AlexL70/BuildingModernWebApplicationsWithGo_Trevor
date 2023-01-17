package main

import (
	"log"
)

type myStruct struct {
	FirstName string
}

func (s *myStruct) printFirstName() string {
	return s.FirstName
}

func main() {
	var myVar myStruct
	myVar.FirstName = "John"
	myVar2 := myStruct{FirstName: "Mary"}
	log.Println("myVar is set to", myVar.printFirstName())
	log.Println("myVar2 is set to", myVar2.printFirstName())
}
