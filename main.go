package main

import (
	"fmt"

	"github.com/nathan-barry/query-quarry/search"
)

func main() {
	fmt.Println("Hello world!")

	filename := "data/wiki40b.test"

	count := search.CountOccurrences(filename, " on Tuesday")
	fmt.Println("count:", count)

	// http.HandleFunc("POST /test", func(w http.ResponseWriter, r *http.Request) {
	// 	if r.Body == nil {
	// 		fmt.Println("ERROR: BODY IS NIL")
	// 	}

	// 	r.Body
	// })

	// http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Println("Test!!!")
	// })

	// log.Fatal(http.ListenAndServe(":8080", nil))
}
