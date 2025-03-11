package api

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// urls to primary key
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
func GenerateURL(primaryKeyFields []string) string {
	var urlParts []string
	for _, field := range primaryKeyFields {
		urlParts = append(urlParts, ":"+ToSnakeCase(strings.ToLower(field)))
	}
	return "/" + strings.Join(urlParts, "/")
}
func GetPrimaryKeyFields[T any](obj T) []string {
	var primaryKeyFields []string
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	max := t.NumField()
	if max == 0 {
		return primaryKeyFields
	}
	for i := range max {
		field := t.Field(i)
		if tag := field.Tag.Get("gorm"); strings.Contains(tag, "primaryKey") {
			primaryKeyFields = append(primaryKeyFields, field.Name)
		}
	}
	if len(primaryKeyFields) == 0 {
		first := v.Field(0)
		if first.Kind() == reflect.Struct {
			primaryKeyFields = GetPrimaryKeyFields(first.Interface())
		} else {
			primaryKeyFields = append(primaryKeyFields, t.Field(0).Name)
		}
	}
	return primaryKeyFields
}

// this is using the url from the configuration of the server
// (so not defined in a http request)
func UrlToQuery(c *gin.Context) (ret string) {
	params := c.Params
	paramValues := make([]string, len(params))
	for i, param := range params {
		paramValues[i] = fmt.Sprintf("%s=?", param.Value)
	}
	ret = strings.Join(paramValues, " AND ")
	return
}

func UrlToValue(c *gin.Context) []interface{} {
	params := c.Params
	paramValues := make([]interface{}, len(params))
	for i, param := range params {
		paramValues[i] = param.Value
	}
	return paramValues
}
