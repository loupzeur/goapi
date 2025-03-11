package api

import (
	"goapi/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

/* Repository to gin handlers */
type RepositoryToGin[T any, Z repositories.Repository[T]] struct {
	BaseUrl    string
	Repository repositories.Repository[T]
}

func (r *RepositoryToGin[T, Z]) SetCRUDRestAPI(g *gin.RouterGroup) {
	pkUrl := GenerateURL(GetPrimaryKeyFields(new(T)))
	g.
		GET(r.BaseUrl, r.GetAll).
		GET(r.BaseUrl+pkUrl, r.GetById).
		POST(r.BaseUrl, r.Post).
		PUT(r.BaseUrl+pkUrl, r.Post).
		PATCH(r.BaseUrl+pkUrl, r.Patch).
		DELETE(r.BaseUrl+pkUrl, r.Delete)
	return
}
func (r *RepositoryToGin[T, Z]) GetAll(ctx *gin.Context) {
	data, err := r.Repository.FindAll(ctx.Request.Context())
	if err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (r *RepositoryToGin[T, Z]) GetById(ctx *gin.Context) {
	data, err := r.Repository.FindByID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (r *RepositoryToGin[T, Z]) Post(ctx *gin.Context) {
	data := new(T)
	if err := ctx.BindJSON(data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Unable to decode request")
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := r.Repository.Create(ctx.Request.Context(), data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (r *RepositoryToGin[T, Z]) Patch(ctx *gin.Context) {
	data := new(T)
	if err := ctx.BindJSON(data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Unable to decode request")
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := r.Repository.Update(ctx.Request.Context(), data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (r *RepositoryToGin[T, Z]) Delete(ctx *gin.Context) {
	data := new(T)
	data, err := r.Repository.FindByID(ctx.Request.Context(), ctx.Param("id"))
	if err = r.Repository.Delete(ctx.Request.Context(), data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
