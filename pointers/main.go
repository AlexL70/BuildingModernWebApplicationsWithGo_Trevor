package main

import "fmt"

func main() {
	var myString string
	myString = "green"

	fmt.Printf("myString is set to %q\n", myString)
	changeUsingPointer(&myString)
	fmt.Printf("myString is now set to %q\n", myString)
}

func changeUsingPointer(p *string) {
	newValue := "red"
	*p = newValue
}
