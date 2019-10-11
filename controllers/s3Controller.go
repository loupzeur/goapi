package controllers

import (
	"fmt"
	"github.com/loupzeur/goapi/models"
	u "github.com/loupzeur/goapi/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/qor/oss/s3"
)

var e = godotenv.Load()
var S3Access = os.Getenv("s3_access_key_id")
var S3Secret = os.Getenv("s3_secret_id")
var S3Region = os.Getenv("s3_region")
var S3Bucket = os.Getenv("s3_bucket")
var S3Url = os.Getenv("s3_url")
var S3Static = os.Getenv("s3_static")
var storage = s3.New(&s3.Config{AccessID: S3Access, AccessKey: S3Secret, Region: S3Region, Bucket: S3Bucket, S3Endpoint: S3Url})

//RegisterS3Route Return routes for this controller
func RegisterS3Route() u.Routes {
	return u.Routes{
		u.Route{"GetAllObject", "GET", "/api/object", GetAllObject},
		u.Route{"GetObject", "GET", "/api/object/{id:[0-9]+}", GetObject},
		u.Route{"PostObject", "POST", "/api/object", PostObject},
		u.Route{"DeleteObject", "DELETE", "/api/object/{id:[0-9]+}", GetObject},
	}
}

//GetObject return the object frm db
func GetObject(w http.ResponseWriter, r *http.Request) {
	models.GenericGet(w, r, &models.Object{}, func(r *http.Request, data interface{}) bool {
		return true
	})
}

//PostObject controller to create images
func PostObject(w http.ResponseWriter, r *http.Request) {
	auth, ok := u.GetAuthenticatedToken(r)
	if !ok {
		u.RespondCode(w, u.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("object")
	if err != nil {
		u.RespondCode(w, u.Message(false, err.Error()), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filepath := "public/"
	filename := fmt.Sprintf("%d-%d-%s", auth.UserId, time.Now().Unix(), handler.Filename)

	o, e := storage.Put(filepath+filename, file)

	if e != nil {
		u.RespondCode(w, u.Message(false, e.Error()), http.StatusInternalServerError)
		return
	}

	obj := &models.Object{
		IDOwner:      auth.UserId,
		Filename:     filename,
		Path:         filepath,
		Size:         handler.Size,
		Type:         handler.Header.Get("Content-Type"),
		DateCreation: time.Now(),
	}

	models.GetDB().Save(obj)

	msg := u.Message(true, "Object Created Successufly")
	msg["filename"] = filename
	msg["url"] = S3Static + filepath + filename
	msg["url_path"] = o.Path
	msg["url_name"] = o.Name
	msg["url_host"] = S3Static
	msg["data"] = obj
	u.Respond(w, msg)
}

//DeleteObject to remove object from db and cdn
func DeleteObject(w http.ResponseWriter, r *http.Request) {
	models.GenericDelete(w, r, &models.Object{}, func(r *http.Request, data interface{}) bool {
		auth, ok := u.GetAuthenticatedToken(r)
		d, tok := data.(*models.Object)
		if !ok || !tok || auth.UserId != d.IDOwner {
			return false
		}
		err := storage.Delete(d.GetPath())
		log.Printf("Error deleting object : %s", err.Error())
		return err == nil
	})
}

//GetAllObject return object from S3
func GetAllObject(w http.ResponseWriter, r *http.Request) {
	msg := u.Message(true, "All Object on bucket")

	list, _ := storage.List("/public/")

	ret := []map[string]string{}
	for _, v := range list {
		ret = append(ret, map[string]string{"Path": v.Path, "Name": v.Name})
	}

	msg["data"] = ret
	u.Respond(w, msg)
}
