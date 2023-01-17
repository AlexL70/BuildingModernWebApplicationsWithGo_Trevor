package main

import (
	"log"
	"sort"
)

type User struct {
	FirstName string
	LastName  string
}

func main() {
	myMap := make(map[string]string)
	myMap["dog"] = "Samson"
	myMap["other-dog"] = "Cassie"
	myMap["dog"] = "Fido"
	log.Println(myMap["dog"])
	log.Println(myMap["other-dog"])

	myIntMap := make(map[string]int)
	myIntMap["First"] = 1
	myIntMap["Second"] = 2
	log.Println(myIntMap["First"], myIntMap["Second"])

	myUserMap := make(map[string]User)
	myUserMap["me"] = User{FirstName: "Alex", LastName: "Levinson"}
	log.Println(myUserMap["me"].FirstName)

	var myStrings []string
	myStrings = append(myStrings, "Alex")
	myStrings = append(myStrings, "John")
	myStrings = append(myStrings, "Mary")
	log.Println(myStrings)
	var myInts []int
	myInts = append(myInts, 2)
	myInts = append(myInts, 1)
	myInts = append(myInts, 3)
	log.Println(myInts)
	sort.Ints(myInts)
	log.Println(myInts)

	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	log.Println(numbers)
	log.Println(numbers[0:2])
	log.Println(numbers[7:9])

	names := []string{"one", "seven", "fish", "cat"}
	log.Println(names)
}
