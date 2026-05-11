package main

import (
	"fmt"
	"log"
	"net/http"
)

type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Finished    bool   `json:"finished"`
	Description string `json:"description"`
}

func main() {
	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":8888", nil))

}
