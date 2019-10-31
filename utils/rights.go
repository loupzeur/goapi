package utils

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

/*
Token JWT claims struct
*/
type Token struct {
	UserId uint
	jwt.StandardClaims
}

//GetAuthenticatedToken return the Authentication token if exists in context
func GetAuthenticatedToken(r *http.Request) (Token, bool) {
	user, ok := r.Context().Value("user").(Token)
	return user, ok
}
