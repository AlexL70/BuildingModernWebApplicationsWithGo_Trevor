package main

import (
	"fmt"
	"net/http"
)

const portNumber = ":8080"

// Home is the home page handler
func Home(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "This is a home page")
}

// About is the about page handler
func About(w http.ResponseWriter, r *http.Request) {
	sum := addValues(3, 2)
	_, _ = fmt.Fprintf(w, "This is an about page. And 3 + 2 is %d", sum)
}

// addValues adds two integers and returns the sum
func addValues(x, y int) int {
	return x + y
}

// main is the main application function
func main() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/about", About)

	fmt.Printf("Starting Web on port %s\n", portNumber)
	_ = http.ListenAndServe(portNumber, nil)
}
