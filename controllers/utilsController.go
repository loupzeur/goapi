package controllers

import (
	"goapi/models"
	u "goapi/utils"
	"net/http"

	"github.com/jinzhu/gorm"
)

func RegisterUtilsRoute() u.Routes {
	return models.CrudRoutes(
		&models.Notification{}, &models.Notification{},
		func(r *http.Request, req *gorm.DB) *gorm.DB {
			auth, ok := u.GetAuthenticatedToken(r)
			if !ok {
				return req.Where("id_user=-1") // to return nothing
			}
			return req.Where("id_user=?", auth.UserId)
		}, u.ReadNotification,
		func(r *http.Request, data interface{}) bool {
			auth, ok := u.GetAuthenticatedToken(r)
			d, tok := data.(*models.Notification)
			return ok && tok && d.IDUser == auth.UserId
		}, u.ReadNotification,
		DefaultRightAccess, u.NotDefined,
		func(r *http.Request, data interface{}, data2 interface{}) bool {
			auth, ok := u.GetAuthenticatedToken(r)
			d, tok := data.(*models.Notification)
			d2, tok2 := data2.(*models.Notification)
			return ok && tok && tok2 && d.IDUser == auth.UserId && d2.IDUser == d.IDUser
		}, u.ReadNotification,
		func(r *http.Request, data interface{}) bool {
			auth, ok := u.GetAuthenticatedToken(r)
			d, tok := data.(*models.Notification)
			return ok && tok && d.IDUser == auth.UserId
		}, u.ReadNotification,
	)
}

//DefaultQueryAll default request for GetAll
func DefaultQueryAll(r *http.Request, req *gorm.DB) *gorm.DB {
	return req
}

//DefaultRightAccess return true right handler
func DefaultRightAccess(r *http.Request, data interface{}) bool {
	return true
}

//DefaultRightEdit return true right handler
func DefaultRightEdit(r *http.Request, data interface{}, data2 interface{}) bool {
	return true
}
