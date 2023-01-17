package main

import (
	"errors"
	"log"
)

func main() {
	result, err := divide(1000.0, 10.0)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Result of division is %f", result)
}

func divide(x, y float32) (float32, error) {
	var result float32
	if y == 0 {
		return result, errors.New("division by zero")
	}
	result = x / y
	return result, nil
}
