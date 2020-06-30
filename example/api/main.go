package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("MESSAGE:", os.Getenv("MESSAGE"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			fmt.Printf("method=%s path=%s took=%s\n", r.Method, r.URL.Path, time.Since(start))
		}(time.Now())

		w.Header().Set("Content-Type", "application/json")

		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			return
		}

		fmt.Fprintf(w, `{"message":"hello"}`)
	})

	port := ":" + os.Getenv("PORT")
	fmt.Printf("Listening on %s...\n", port)
	fmt.Println(http.ListenAndServe(port, nil))
}
