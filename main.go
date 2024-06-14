package main

import (
	"fmt"
	"time"

	"github.com/nathan-barry/query-quarry/search"
)

func main() {
	t := time.Now()

	filename := "data/wiki40b.test"

	count := search.CountOccurrences(filename, " on Tuesday")
	fmt.Println("count:", count)
	fmt.Println("Time taken:", time.Since(t).Seconds())

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
