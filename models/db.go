package models

//https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/loupzeur/goapi/utils"

	"github.com/jinzhu/gorm"
	"github.com/smallnest/gen/dbmeta"

	//To retrieve mysql functions
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
)

//Validation interface to validate stuff
type Validation interface {
	TableName() string
	Validate() (map[string]interface{}, bool)
	OrderColumns() []string
	FilterColumns() map[string]string
	Authorization(r *http.Request) (bool, string)
}

var db *gorm.DB //database

func init() {
	e := godotenv.Load() //Load .env file
	if e != nil {
		fmt.Println(e)
	}
	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")
	dbPort := os.Getenv("db_port")
	if dbPort == "" {
		dbPort = "3306"
	}

	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", username, password, dbHost, dbPort, dbName)
	conn, err := gorm.Open("mysql", dbURI)
	if err != nil {
		fmt.Println(err)
	}
	db = conn
}

//GetDB returns a handle to the DB object
func GetDB() *gorm.DB {
	return db
}

//CrudRoutesSpecificURL Generate default CRUD route for object with specific url
func CrudRoutesSpecificURL(models Validation,
	freq func(r *http.Request, req *gorm.DB) *gorm.DB,
	getfunc func(r *http.Request, data interface{}) bool,
	crefunc func(r *http.Request, data interface{}) bool,
	updfunc func(r *http.Request, data interface{}, data2 interface{}) bool,
	delfunc func(r *http.Request, data interface{}) bool, url string) utils.Routes {
	return utils.Routes{
		utils.Route{"GetAll" + strings.Title(models.TableName()), "GET", "/api/" + url + models.TableName(),
			func(w http.ResponseWriter, r *http.Request) {
				GenericGetQueryAll(w, r, models, freq)
			}},
		utils.Route{"Get" + strings.Title(models.TableName()), "GET", "/api/" + url + models.TableName() + "/{id:[0-9]+}",
			func(w http.ResponseWriter, r *http.Request) {
				GenericGet(w, r, models, getfunc)
			}},
		utils.Route{"Create" + strings.Title(models.TableName()), "POST", "/api/" + url + models.TableName(),
			func(w http.ResponseWriter, r *http.Request) {
				GenericCreate(w, r, models, crefunc)
			}},
		utils.Route{"Update" + strings.Title(models.TableName()), "PUT", "/api/" + url + models.TableName() + "/{id:[0-9]+}",
			func(w http.ResponseWriter, r *http.Request) {
				GenericUpdate(w, r, models, updfunc)
			}},
		utils.Route{"Delete" + strings.Title(models.TableName()), "DELETE", "/api/" + url + models.TableName() + "/{id:[0-9]+}",
			func(w http.ResponseWriter, r *http.Request) {
				GenericDelete(w, r, models, delfunc)
			}},
	}
}

//CrudRoutes Generate default CRUD route for object
func CrudRoutes(models Validation,
	freq func(r *http.Request, req *gorm.DB) *gorm.DB,
	getfunc func(r *http.Request, data interface{}) bool,
	crefunc func(r *http.Request, data interface{}) bool,
	updfunc func(r *http.Request, data interface{}, data2 interface{}) bool,
	delfunc func(r *http.Request, data interface{}) bool) utils.Routes {
	return CrudRoutesSpecificURL(models,
		freq,
		getfunc,
		crefunc,
		updfunc,
		delfunc, "")
}

//GetAllFromDb return paginated database
func GetAllFromDb(r *http.Request) (int64, int64, string) {
	page, err := utils.ReadInt(r, "page", 1)
	if err != nil || page < 1 {
		return 0, 0, ""
	}
	pagesize, err := utils.ReadInt(r, "pagesize", 20)
	if err != nil || pagesize <= 0 {
		return 0, 0, ""
	}
	offset := (page - 1) * pagesize
	order := r.FormValue("order")
	return offset, pagesize, order
}

//GenericGetQueryAll return all elements with filters
func GenericGetQueryAll(w http.ResponseWriter, r *http.Request, data Validation, freq func(r *http.Request, req *gorm.DB) *gorm.DB) {
	dtype := reflect.TypeOf(data)
	pages := reflect.New(reflect.SliceOf(dtype)).Interface()
	//Limit and Pagination Part
	offset, pagesize, order := GetAllFromDb(r)
	err := error(nil)
	if offset <= 0 && pagesize <= 0 {
		err = errors.New("error with elements size")
	}
	//Ordering Part
	hasOrders := false //avoid sql injection on orders
	for _, v := range data.OrderColumns() {
		val := strings.Split(order, "_")
		orderDirection := val[len(val)-1]
		if len(val) >= 2 && strings.HasPrefix(order, v) && (orderDirection == "asc" || orderDirection == "desc") {
			hasOrders = true
			order = v + " " + strings.ToUpper(orderDirection)
			break
		}
	}
	if !hasOrders {
		order = ""
	}
	req := GetDB().LogMode(true).Set("gorm:auto_preload", true).Model(data)

	//Get Default Query
	req = freq(r, req)

	if order != "" {
		req = req.Order(order)
	}
	//Additionnal Querying Part
	urlvars := r.URL.Query()
	//Remove useless parameters to avoid iterating over filters for nothing ^^
	delete(urlvars, "page")
	delete(urlvars, "order")
	delete(urlvars, "pagesize")
	if len(urlvars) > 0 {
		for k, v := range data.FilterColumns() {
			if val, ok := urlvars[k]; ok {
				switch v {
				case "in":
					req = req.Where(k+" IN (?)", val)
				case "stringlike":
					req = req.Where(k+" LIKE ?", "%"+val[0]+"%")
				default:
					req = req.Where(k+"=?", val[0])
				}

			}
		}
	}

	//Execution request Part
	count := 0
	err = req.Count(&count).Error
	err = req.Offset(offset).Limit(pagesize).Find(pages).Error
	if err != nil {
		utils.Respond(w, utils.Message(false, "Error while retrieving data"))
		return
	}

	resp := utils.Message(true, "data returned")
	resp["data"] = pages
	resp["total_nb_values"] = count
	resp["current_page"] = offset/pagesize + 1
	resp["size_page"] = pagesize
	utils.Respond(w, resp)
}

//Controllers Generic Accessors

//GenericGet default controller for get
func GenericGet(w http.ResponseWriter, r *http.Request, data interface{}, f func(r *http.Request, data interface{}) bool) {
	tmp := reflect.New(reflect.TypeOf(data).Elem()).Interface()
	err := GetFromID(r, tmp)
	if !f(r, tmp) {
		err = errors.New("Access Forbidden")
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Not Found"), http.StatusNotFound)
		return
	}
	resp := utils.Message(true, "success")
	resp["data"] = tmp
	utils.Respond(w, resp)
}

//GenericCreate create a new object
func GenericCreate(w http.ResponseWriter, r *http.Request, data Validation, f ...func(r *http.Request, data interface{}) bool) {
	tmp := reflect.New(reflect.TypeOf(data).Elem()).Interface().(Validation)
	err := createFromJSONRequest(r, tmp)
	actions := len(f)
	reason, ok := tmp.Validate()
	if !ok {
		utils.RespondCode(w, reason, http.StatusNotAcceptable)
		return
	}

	if actions > 0 && !f[0](r, tmp) {
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err = GetDB().Save(tmp).Error; err != nil {
		utils.RespondCode(w, utils.Message(false, "Error saving"), http.StatusInternalServerError)
		return
	}
	if actions == 2 {
		f[1](r, tmp) //notification, ...
	}
	resp := utils.Message(true, "success")
	resp["data"] = tmp
	utils.Respond(w, resp)
}

//GenericUpdate default updater for controller
func GenericUpdate(w http.ResponseWriter, r *http.Request, data Validation, f func(r *http.Request, data interface{}, data2 interface{}) bool) {
	tmp1 := reflect.New(reflect.TypeOf(data).Elem()).Interface().(Validation)
	tmp2 := reflect.New(reflect.TypeOf(data).Elem()).Interface()

	err := updateFromID(r, tmp1, tmp2)
	val, ret := tmp1.Validate()
	if !ret {
		utils.RespondCode(w, val, http.StatusNotAcceptable)
		return
	}
	if !f(r, tmp1, tmp2) {
		err = errors.New("Access Forbidden")
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err := dbmeta.Copy(tmp1, tmp2); err != nil {
		utils.RespondCode(w, utils.Message(false, "Data Error"), http.StatusInternalServerError)
		return
	}
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Not Found"), http.StatusNotFound)
		return
	}
	if err = GetDB().Save(tmp1).Error; err != nil {
		utils.RespondCode(w, utils.Message(false, "Error saving"), http.StatusInternalServerError)
		return
	}
	resp := utils.Message(true, "success")
	resp["data"] = tmp1
	utils.Respond(w, resp)
}

//GenericDelete default deleter for controller
func GenericDelete(w http.ResponseWriter, r *http.Request, data interface{}, f func(r *http.Request, data interface{}) bool) {
	tmp := reflect.New(reflect.TypeOf(data).Elem()).Interface()
	err := deleteFromID(r, tmp)
	if !f(r, tmp) {
		err = errors.New("Access Forbidden")
		utils.RespondCode(w, utils.Message(false, "Forbidden"), http.StatusForbidden)
		return
	}
	if err != nil {
		utils.RespondCode(w, utils.Message(false, "Not Found"), http.StatusNotFound)
		return
	}
	if err = GetDB().Delete(tmp).Error; err != nil {
		utils.RespondCode(w, utils.Message(false, "Error saving"), http.StatusInternalServerError)
		return
	}
	utils.Respond(w, utils.Message(true, "Deletion successful"))
}

//Internals

//Generic Functions for CRUD

func createFromJSONRequest(r *http.Request, data interface{}) error {
	if err := utils.ReadJSON(r, data); err != nil {
		return err
	}
	return nil
}

func deleteFromID(r *http.Request, data interface{}) error {
	id, err := utils.ReadIntURL(r, "id")
	if err != nil {
		return err
	}
	if err := GetDB().First(data, id).Error; err != nil {
		return err
	}
	return nil
}

//GetFromID Return object from Id
func GetFromID(r *http.Request, data interface{}) error {
	id, err := utils.ReadIntURL(r, "id")
	if err != nil {
		return err
	}
	if err := GetDB().Set("gorm:auto_preload", true).First(data, id).Error; err != nil {
		return err
	}
	return nil
}

func updateFromID(r *http.Request, data1 interface{}, data2 interface{}) error {
	id, err := utils.ReadIntURL(r, "id")
	if err != nil {
		return err
	}
	if err := GetDB().First(data1, id).Error; err != nil {
		return err
	}
	if err := utils.ReadJSON(r, data2); err != nil {
		return err
	}
	return nil
}
