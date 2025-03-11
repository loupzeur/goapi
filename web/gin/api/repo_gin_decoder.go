package api

import (
	"goapi/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// management of POST, PUT and PATCH
// Decodable POST, PUT and PATCH
type BodyDecoder[T any] interface {
	To() *T
}
type DecodableGin[
	T any,
	POST BodyDecoder[T],
	PUT BodyDecoder[T],
	PATCH BodyDecoder[T],
] struct {
	RepositoryToGin[T, repositories.Repository[T]]
}

func NewGin[T any, POST, PUT, PATCH BodyDecoder[T]](repo repositories.Repository[T]) *DecodableGin[T, POST, PUT, PATCH] {
	return &DecodableGin[T, POST, PUT, PATCH]{
		RepositoryToGin: RepositoryToGin[T, repositories.Repository[T]]{
			Repository: repo,
		},
	}
}
func (d *DecodableGin[T, POST, PUT, PATCH]) Post(ctx *gin.Context) {
	data, err := bodyDecoderFunc[T, POST](ctx)
	if err != nil {
		log.Ctx(ctx.Request.Context()).Error().Interface("Body", data).Err(err).Msg("Unable to decode request")
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := d.Repository.Create(ctx.Request.Context(), data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (d *DecodableGin[T, POST, PUT, PATCH]) Put(ctx *gin.Context) {
	data, err := bodyDecoderFunc[T, PUT](ctx)
	if err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Unable to decode request")
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := d.Repository.Create(ctx.Request.Context(), data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (d *DecodableGin[T, POST, PUT, PATCH]) Patch(ctx *gin.Context) {
	data, err := bodyDecoderFunc[T, POST](ctx)
	if err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Unable to decode request")
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := d.Repository.Create(ctx.Request.Context(), data); err != nil {
		log.Ctx(ctx.Request.Context()).Err(err).Msg("Get all error")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (d *DecodableGin[T, POST, PUT, PATCH]) SetCRUDRestAPI(g *gin.RouterGroup) {
	pkUrl := GenerateURL(GetPrimaryKeyFields(new(T)))
	g.
		GET(d.BaseUrl, d.GetAll).
		GET(d.BaseUrl+pkUrl, d.GetById).
		POST(d.BaseUrl, d.Post).
		PUT(d.BaseUrl+pkUrl, d.Put).
		PATCH(d.BaseUrl+pkUrl, d.Patch).
		DELETE(d.BaseUrl+pkUrl, d.Delete)
	return
}
func bodyDecoderFunc[X any, Y BodyDecoder[X]](ctx *gin.Context) (*X, error) {
	data := new(Y)
	if err := ctx.BindJSON(data); err != nil {
		return nil, err
	}
	return (*data).To(), nil
}
