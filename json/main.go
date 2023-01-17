package main

import (
	"encoding/json"
	"log"
)

type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	HairColor string `json:"hair_color"`
	HasDog    bool   `json:"has_dog"`
}

func main() {
	myJson := `
[
	{
		"first_name": "Clark",
		"last_name": "Kent",
		"hair_color": "black",
		"has_dog": true
	},
	{
		"first_name": "Bruce",
		"last_name": "Wayne",
		"hair_color": "black",
		"has_dog": false
	}
]
`
	var unmarshalled []Person
	err := json.Unmarshal([]byte(myJson), &unmarshalled)
	if err != nil {
		log.Printf("Error unmarshalling json: %s\n", err)
		return
	}

	log.Printf("Unmarshalled: %v", unmarshalled)

	//	write json from a struct
	var mySlice = []Person{
		Person{"Wally", "West", "red", false},
		Person{"Diana", "Prince", "black", false},
	}
	jsonStr, err := json.MarshalIndent(mySlice, "", "    ")
	if err != nil {
		log.Printf("Error marshalling json: %s\n", err)
		return
	}
	log.Printf("Marshalled: %s\n", string(jsonStr))
}
