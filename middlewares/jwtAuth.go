package middlewares

import (
	"context"
	u "goapi/utils"
	"net/http"
	"os"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	//"fmt"
)

//Routes from main
var Routes u.Routes

//JwtAuthentication jwt auth checker handler -> set token for user
func JwtAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		response := make(map[string]interface{})
		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader != "" {
			splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
			if len(splitted) != 2 {
				response = u.Message(false, "Invalid/Malformed auth token")
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("Content-Type", "application/json")
				u.Respond(w, response)
				return
			}

			tokenPart := splitted[1] //Grab the token part, what we are truly interested in
			tk := &u.Token{}

			token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("token_password")), nil
			})

			if err != nil { //Malformed token, returns with http code 403 as usual
				response = u.Message(false, "Malformed authentication token")
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("Content-Type", "application/json")
				u.Respond(w, response)
				return
			}

			if !token.Valid { //Token is invalid, maybe not signed on this server
				response = u.Message(false, "Token is not valid.")
				w.WriteHeader(http.StatusForbidden)
				w.Header().Add("Content-Type", "application/json")
				u.Respond(w, response)
				return
			}

			//Everything went well, proceed with the request and set the caller to the user retrieved from the parsed token
			ctx := context.WithValue(r.Context(), "user", *tk)
			r = r.WithContext(ctx)
		}//Else ... no auth

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
