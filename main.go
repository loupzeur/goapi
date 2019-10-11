package goapi

import (
	"log"
	"net/http"
	"os"
	"github.com/loupzeur/goapi/controllers"
	"github.com/loupzeur/goapi/middlewares"
)

//Usage exemple
func main() {
	middlewares.Routes = controllers.RegisterUserRoute().
		Append(controllers.RegisterS3Route())

	router := NewRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Fatal(http.ListenAndServe(":"+port, router))

}
