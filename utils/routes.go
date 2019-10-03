package utils

import (
	"net/http"
)

//Route define a route with url and rights required to access it
type Route struct {
	Name          string
	Method        string
	Pattern       string
	HandlerFunc   http.HandlerFunc
	Authorization uint32
}

//Routes an array of route
type Routes []Route

//Append Add routes to Routes
func (r Routes) Append(routes []Route) Routes {
	r = append(r, routes...)
	return r
}
