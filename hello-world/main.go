package main

import (
	"errors"
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

func Divide(w http.ResponseWriter, r *http.Request) {
	f, err := divideValues(100.0, 10.0)
	if err != nil {
		fmt.Fprint(w, "Cannot divide by zero")
		return
	}
	fmt.Fprintf(w, "%f / %f = %f", 100.0, 10.0, f)
}

func divideValues(x, y float32) (float32, error) {
	if y == 0.0 {
		return 0.0, errors.New("divide by zero")
	}
	return x / y, nil
}

// main is the main application function
func main() {
	http.HandleFunc("/", Home)
	http.HandleFunc("/about", About)
	http.HandleFunc("/divide", Divide)

	fmt.Printf("Starting Web on port %s\n", portNumber)
	_ = http.ListenAndServe(portNumber, nil)
}
