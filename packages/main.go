package main

import (
	"log"

	"github.com/AlexL70/myniceprogram/helpers"
)

func main() {
	var myVar helpers.SomeType
	myVar.TypeName = "My Type"
	myVar.TypeNumber = 11
	log.Println(myVar)
}
