package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		n, err := fmt.Fprintf(w, "Hello, Web!")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Number of bytes written is %d\n", n)
	})
	_ = http.ListenAndServe(":8080", nil)
}
