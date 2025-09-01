package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from backend!")
	})
	fmt.Println("Dummy backend running on :9000")
	http.ListenAndServe(":9000", nil)
}
