package main

import (
	"fmt"
	"log"
)

type User struct {
	FirstName string
	LastName  string
	Email     string
	Age       int
}

func main() {
	for i := 0; i < 10; i++ {
		log.Println(i)
	}
	log.Println("----------------------------------------")
	animals := []string{"dog", "fish", "horse", "cow", "cat"}
	for i, animal := range animals {
		log.Println(i, animal)
	}
	log.Println("----------------------------------------")
	animal_map := map[string]string{
		"dog": "Fido",
		"cat": "Fluffy",
	}
	for animalType, animal := range animal_map {
		log.Println(animalType, animal)
	}
	log.Println("----------------------------------------")
	var firstLine = "Once upon a midnight dreary"
	for i, c := range firstLine {
		fmt.Print(i, ":", c, "   ")
	}
	fmt.Println()
	log.Println("----------------------------------------")
	var users = []User{
		User{"John", "Smith", "john.smith@example.com", 33},
		User{"Mary", "Jones", "mary.jones@example.com", 51},
		User{"Sally", "Brown", "sally.brown@example.com", 27},
		User{"Alex", "Anderson", "alex.anderson@example.com", 64},
	}
	for _, user := range users {
		log.Println(user.FirstName, user.LastName, user.Email, user.Age)
	}
}
