package api_test

import (
	"goapi/repositories"
	ginapi "goapi/web/gin/api"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dbObject struct {
	gorm.Model
	Name string
}

type dbObjectRepo struct {
	repositories.Repository[dbObject]
}

func NewRepo(db *gorm.DB) *dbObjectRepo {
	return &dbObjectRepo{
		Repository: repositories.NewRepository[dbObject](db),
	}
}

func TestHandlers(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	db.AutoMigrate(&dbObject{})

	nr := NewRepo(db)

	g := ginapi.RepositoryToGin[dbObject, repositories.Repository[dbObject]]{
		BaseUrl:    "",
		Repository: nr,
	}

	assert.Equal(t, g.BaseUrl, "test")
}

func TestUrl(t *testing.T) {
	pkUrl := ginapi.GenerateURL(ginapi.GetPrimaryKeyFields(new(dbObject)))
	t.Log(pkUrl)
}
