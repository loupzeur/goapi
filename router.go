package goapi

import (
	"net/http"
	"github.com/loupzeur/goapi/middlewares"

	"github.com/gorilla/mux"
)

//NewRouter set router and activate required handlers
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	router.Use(middlewares.JwtAuthentication)
	for _, route := range middlewares.Routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = middlewares.Logger(handler, route.Name)
		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}
	return router
}
