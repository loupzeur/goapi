package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// this one doesn't use the repository
type ScopableGet interface {
	QueryGet(*gorm.DB) *gorm.DB
}
type ScopablePost interface {
	QueryPost(context.Context) func(*gorm.DB) *gorm.DB
}
type ScopablePatch interface {
	QueryPost(context.Context) func(*gorm.DB) *gorm.DB
	QueryGet(*gorm.DB) *gorm.DB
}
type ScopableDelete interface {
	QueryDelete(context.Context) func(*gorm.DB) *gorm.DB
}

type Config struct {
	Port   string
	routes map[string]http.HandlerFunc
	server *http.ServeMux
}

func NewAPI(port string) *Config {
	return &Config{
		Port:   port,
		routes: map[string]http.HandlerFunc{},
		server: http.NewServeMux(),
	}
}

func (c *Config) Config(cfgs ...func(c *Config)) {
	for _, f := range cfgs {
		f(c)
	}
	for path, handler := range c.routes {
		c.server.HandleFunc(path, handler)
	}
}
func (c *Config) Serve() error {
	t := http.Server{
		Addr:    c.Port,
		Handler: c.server,
	}
	return t.ListenAndServe()
}
func WithRoute(url string, f http.HandlerFunc) func(c *Config) {
	return func(c *Config) {
		c.routes[url] = f
	}
}

// utilities functions

func Get[T any](db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new([]T)
		if err := db.WithContext(r.Context()).Find(data).Error; err != nil {
			log.Error().Ctx(r.Context()).Err(err).Msg("unable to retrieve data")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Debug().Ctx(r.Context()).Err(json.NewEncoder(w).Encode(data)).Msg("Encoding response GET")
	}
}
func GetByField[T ScopableGet](db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new(T)
		if err := db.WithContext(r.Context()).Scopes((*data).QueryGet).First(data, queryFromPath(r)...).Error; err != nil {
			log.Error().Ctx(r.Context()).Err(err).Msg("unable to retrieve data")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Debug().Ctx(r.Context()).Err(json.NewEncoder(w).Encode(data)).Msg("Encoding response getbyfield")
	}
}
func Post[T ScopablePost](db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new(T)
		log.Debug().Ctx(r.Context()).Err(json.NewDecoder(r.Body).Decode(data)).Msg("Decoding POST query")
		if err := db.WithContext(r.Context()).Scopes((*data).QueryPost(r.Context())).Save(data).Error; err != nil {
			log.Error().Ctx(r.Context()).Err(err).Msg("unable to save data")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Debug().Ctx(r.Context()).Err(json.NewEncoder(w).Encode(data)).Msg("Encoding POST response")
	}
}
func Patch[T ScopablePatch](db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new(T)
		log.Debug().Ctx(r.Context()).Err(json.NewDecoder(r.Body).Decode(data)).Msg("Decoding Patch query")
		if err := db.WithContext(r.Context()).Scopes((*data).QueryPost(r.Context())).Updates(data).Error; err != nil {
			log.Error().Ctx(r.Context()).Err(err).Msg("unable to save data")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := db.WithContext(r.Context()).Scopes((*data).QueryGet).First(data, queryFromPath(r)...).Error; err != nil {
			log.Error().Ctx(r.Context()).Err(err).Msg("unable to retrieve data")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		log.Debug().Ctx(r.Context()).Err(json.NewEncoder(w).Encode(data)).Msg("Encoding POST response")
	}
}
func Delete[T ScopableDelete](db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := new(T)
		req := db.WithContext(r.Context()).Scopes((*data).QueryDelete(r.Context())).Delete(data, queryFromPath(r)...)
		if err := req.Error; err != nil {
			log.Error().Ctx(r.Context()).Err(err).Msg("unable to delete data")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if req.RowsAffected > 0 {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

var vars = regexp.MustCompile(`\{([^}]+)\}`)

// TODO optimise this?
func queryFromPath(r *http.Request) []any {
	params := vars.FindAllString(r.Pattern, -1)
	query, vars := []string{}, []any{}
	for _, match := range params {
		match = strings.ReplaceAll(strings.ReplaceAll(match, "}", ""), "{", "")
		vars = append(vars, r.PathValue(match))
		query = append(query, fmt.Sprintf("%s=?", match))
	}
	return append([]any{strings.Join(query, " AND ")}, vars...)
}

var UserValue = "USERKEY"

func IDMDW(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserValue, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
