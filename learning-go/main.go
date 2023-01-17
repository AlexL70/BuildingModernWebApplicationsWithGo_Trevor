package main

import (
	"fmt"
	"time"
)

var s = "seven"

type User struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Age         int
	BirthDate   time.Time
}

func main() {
	s2 := "six"
	s := "eight"

	fmt.Printf("s is %q\n", s)
	fmt.Printf("s2 is %q\n", s2)

	saySomething("xxx")

	user := User{
		FirstName:   "Trevor",
		LastName:    "Sawler",
		PhoneNumber: "11111111",
		Age:         50,
	}

	fmt.Printf("User Name is %q and last name is %q and birthdate is %q\n", user.FirstName, user.LastName, user.BirthDate)
}

func saySomething(s3 string) (string, string) {
	fmt.Printf("s inside saySomething function is %q\n", s)
	return s3, "World"
}
