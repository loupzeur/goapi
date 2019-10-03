package goapi

import (
	"goapi/controllers"
	"goapi/middlewares"
	"net/http"

	"github.com/gorilla/mux"
)

//NewRouter set router and activate required handlers
func NewRouter() *mux.Router {

	middlewares.Routes = controllers.RegisterUserRoute().
		Append(controllers.RegisterS3Route()).
		Append(controllers.RegisterUtilsRoute())

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
