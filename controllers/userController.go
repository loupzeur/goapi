package controllers

import (
	"encoding/json"
	"goapi/models"
	u "goapi/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

//RegisterUserRoute Return routes for this controller
func RegisterUserRoute() u.Routes {
	return u.Routes{
		u.Route{"GetAllUser", "GET", "/api/user", GetAllUsers},
		u.Route{"GetUser", "GET", "/api/user/{id:[0-9]+}", GetUser},
		u.Route{"CreateUser", "POST", "/api/user", CreateUser},
		u.Route{"UpdateUser", "PUT", "/api/user/{id:[0-9]+}", UpdateUser},
		u.Route{"UpdateUserToken", "PUT", "/api/user/{id:[0-9]+}/token/{token}", UpdateUser},
		u.Route{"DeleteUser", "DELETE", "/api/user/{id:[0-9]+}", DeleteUser},
		//Secondary
		u.Route{"AuthenticateUser", "POST", "/api/user/authenticate", AuthenticateUser},
		u.Route{"RefreshUser", "GET", "/api/user/refresh", RefreshUser},
		u.Route{"ValidateUser", "GET", "/api/user/{id:[0-9]+}/validate", ValidateUser},
		u.Route{"ResetUser", "POST", "/api/user/{email}/reset", ResetUser},
	}
}

//ValidateUser validate a user
func ValidateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		//The passed path parameter is not an integer
		u.Respond(w, u.Message(false, "There was an error in your request"))
		return
	}
	data := models.GetUser(uint(id))
	if data == nil {
		u.RespondCode(w, u.Message(false, "Not Found"), http.StatusNotFound)
		return
	}

	data.Active = 1
	if err := models.GetDB().Save(data).Error; err != nil {
		u.RespondCode(w, u.Message(false, "There was an error while saving"), http.StatusInternalServerError)
		return
	}
	u.Respond(w, u.Message(true, "Activation OK"))
}

//ResetUser add a reset password token
func ResetUser(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}

	email, existing := mux.Vars(r)["email"]

	if !existing {
		u.RespondCode(w, u.Message(false, "User not found"), 404) //... pas sûr qu'une 404 soit bien bruteforçage d'email ?
		return
	}

	err := models.GetDB().Table(user.TableName()).First(&user, "email=?", email).Error

	if err != nil {
		u.RespondCode(w, u.Message(false, "User not found"), 404) //... pas sûr qu'une 404 soit bien bruteforçage d'email ?
		return
	}
	token := u.GetSha([]byte(time.Now().String()))
	user.ResetPasswordToken = token
	user.LastPasswdGen = time.Now()
	models.GetDB().Save(user)

	u.Respond(w, u.Message(err == nil, "Password reset"))
}

//AuthenticateUser return jwt auth token
func AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}
	err := json.NewDecoder(r.Body).Decode(user) //decode the request body into struct and failed if any error occur
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp := models.LoginUser(user.Email, user.Passwd)
	u.Respond(w, resp)
}

//RefreshUser return jwt auth token
func RefreshUser(w http.ResponseWriter, r *http.Request) {
	user := &models.User{}

	auth, ok := u.GetAuthenticatedToken(r)
	err := models.GetDB().Set("gorm:auto_preload", true).Table(user.TableName()).First(user, "id_user=?", auth.UserId).Error
	if !ok || err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp := u.Message(true, "Token Created")
	resp["data"] = user.GenToken()
	u.Respond(w, resp)
}

// ----------------Replace ?-------------------

//GetAllUsers return all users
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	models.GenericGetAll(w, r, &models.User{})
}

//CreateUser creates an account for user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	token := u.GetSha([]byte(time.Now().String()))
	models.GenericCreate(w, r, &models.User{}, func(r *http.Request, data interface{}) bool {
		user, _ := data.(*models.User)
		user.ResetPasswordToken = token
		return true
	})
}

//GetUser return a user
func GetUser(w http.ResponseWriter, r *http.Request) {
	models.GenericGet(w, r, &models.User{}, func(r *http.Request, data interface{}) bool {
		auth, ok := u.GetAuthenticatedToken(r)
		user, valid := data.(*models.User)
		user.Passwd = ""
		return ok && (valid && user.IDUser == auth.UserId)
	})
}

//UpdateUser update a user from db
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	models.GenericUpdate(w, r, &models.User{}, &models.User{}, func(r *http.Request, data interface{}, data2 interface{}) bool {
		auth, ok := u.GetAuthenticatedToken(r)
		user, valid := data.(*models.User)
		updated, _ := data2.(*models.User)
		if user.IDUser != updated.IDUser {
			return false
		}
		if updated.Passwd != "" {
			updated.NewPassword(updated.Passwd)
		}

		token := mux.Vars(r)["token"]

		return (token != "" && token == user.ResetPasswordToken) || //Modification par token (reset/validation du compte)
			(ok && (valid && auth.UserId == user.IDUser)) // Modification par token de session
	})
}

//DeleteUser remove a user from db
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	models.GenericDelete(w, r, &models.User{}, func(r *http.Request, data interface{}) bool {
		_, ok := u.GetAuthenticatedToken(r)
		return ok
	})
}
