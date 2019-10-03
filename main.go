package goapi

import (
	"log"
	"net/http"
	"os"
)

func main() {
	router := NewRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Fatal(http.ListenAndServe(":"+port, router))

}
