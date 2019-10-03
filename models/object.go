package models

import (
	"database/sql"
	"time"

	"github.com/guregu/null"
)

var (
	_ = time.Second
	_ = sql.LevelDefault
	_ = null.Bool{}
)

//Object stored on S3
type Object struct {
	IDObject     int       `gorm:"column:id_object;primary_key" json:"id_object"`
	IDOwner      uint      `gorm:"column:id_owner" json:"id_owner"`
	Filename     string    `gorm:"column:filename" json:"filename"`
	Path         string    `gorm:"column:path" json:"path"`
	Size         int64     `gorm:"column:size" json:"size"`
	Type         string    `gorm:"column:type" json:"type"`
	DateCreation time.Time `gorm:"column:date_creation" json:"date_creation"`
}

// TableName sets the insert table name for this struct type
func (a *Object) TableName() string {
	return "object"
}

//Validate to validate a model
func (a *Object) Validate() (map[string]interface{}, bool) {
	if a.IDObject == 0 {
		a.DateCreation = time.Now()
	}
	return nil, true
}

//OrderColumns return available order columns
func (a *Object) OrderColumns() []string {
	return []string{"date_creation", "filename", "type"}
}

//FilterColumns to return default columns to filter on
func (a *Object) FilterColumns() map[string]string {
	return map[string]string{"id_object": "int"}
}

//GetPath return path of object
func (a *Object) GetPath() string {
	return a.Path + a.Filename
}
