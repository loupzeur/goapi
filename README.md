# Generic Golang APi System
Produce a default API system with a generic controller

## Default controller for get (all, and id), post, put and delete
All models return by that controller must implement models.Validation

### Default usage

Define routes through models.CrudRoutes(

    models models.Validation, 

    freq func(r *http.Request, req *gorm.DB) *gorm.DB, 

    getfunc func(r *http.Request, data interface{}) bool, 

    crefunc func(r *http.Request, data interface{}) bool, 

    updfunc func(r *http.Request, data interface{}, data2 interface{}) bool, 

    delfunc func(r *http.Request, data interface{}) bool) u.Routes

This will create 5 routes :

GetAll, Get, Create, Update and Delete

Using all default function would be :

### Examples
middlewares.Routes = utils.Routes{models.CrudRoutes(

    myObject{}, 

    DefaultQueryAll, 

    DefaultRightAccess, 

    DefaultRightAccess, 

    DefaultRightEdit, 

    DefaultRightAccess)

}

router := NewRouter() // will read routes from middlewares.Routes
log.Fatal(http.ListenAndServe(":8000", router))

## Model
Simply define your model, and implement models.Validation interface :

`
type Validation interface {

	TableName() string

	Validate() (map[string]interface{}, bool)

	OrderColumns() []string

	FilterColumns() map[string]string

	Authorization(r *http.Request) (bool, string)

}
`

### Validation interface

TableName() string Is the name of the table

Validate() is called on post and put call to verify if model is ok

OrderColumns()[]string return an array of column name allowed to be sorted on

FilterColumns() allow filter on column (choose string like, ...)

Authorization() to validate authrorization


# Configuration
.env file or system environment stuff

## Amazon S3 env variable
    s3_access_key_id
    s3_secret_id
    s3_region
    s3_bucket
    s3_url
    s3_static

## Database env variable
    db_user
    db_pass
    db_name
    db_host
    db_port