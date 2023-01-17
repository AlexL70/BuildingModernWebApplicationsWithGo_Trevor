package main

import "log"

func main() {
	var isTrue bool
	isTrue = true
	if isTrue {
		log.Println("isTrue is", isTrue)
	} else {
		log.Print("else isTrue is", isTrue)
	}

	cat := "cat2"
	if cat == "cat" {
		log.Println("Cat is cat")
	} else {
		log.Println("Cat is not cat")
	}

	isTrue = false
	myNum := 100
	if myNum > 99 && isTrue {
		log.Println("myNum is greater than 99 and isTrue is set to true")
	}

	myVar := "bird"
	switch myVar {
	case "cat":
		log.Println("myVar is set to cat")
	case "dog":
		log.Println("myVar is set to dog")
	case "fish":
		log.Println("myVar is set to fish")
	case "whale":
		log.Println("myVar is set to whale")
	default:
		log.Println("myVar is set to", myVar)
	}
}
