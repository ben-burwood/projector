package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message":"Hello, World!"}`))
	})

	mux.Handle("/", http.FileServer(http.Dir("./frontend/dist")))

	log.Fatal(http.ListenAndServe("[::]:8080", mux))
}
