package midw

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
)

var CookieAuthToken = "token"

type JWTConfig struct {
	Key             string `env:"JWT_KEY" envDefault:"DefaultKey"`
	Issuer          string `env:"JWT_ISSUER" envDefault:"NoIssuer"`
	ExpiresIn       int    `env:"JWT_EXPIRES_IN" envDefault:"12"`
	MidwStorage     string `env:"JWT_MIDW_AUTH_STORAGE" envDefault:"subject"`
	UrlAuthRedirect string `env:"JWT_AUTH_REDIRECT" envDefault:""`
	CookieName      string `env:"JWT_COOKIE_NAME" envDefault:"token"`

	//Function that return the token default : DefaultTokenRetriever
	Retriever func(c *gin.Context) string
}

func NewJWTConfig() (*JWTConfig, error) {
	cfg, err := env.ParseAs[JWTConfig]()
	if err != nil {
		return nil, err
	}
	cfg.Retriever = DefaultTokenRetriever(cfg.CookieName)
	return &cfg, nil
}

func DefaultTokenRetriever(cookieName string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		token := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			cookie, err := c.Request.Cookie(CookieAuthToken)
			if err != nil {
				return token
			}
			token = cookie.Value
		}
		return token
	}
}
func (cfg JWTConfig) JWTMiddleware(required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := cfg.Retriever(c)
		if token == "" && required {
			if cfg.UrlAuthRedirect != "" {
				c.Redirect(http.StatusFound, cfg.UrlAuthRedirect)
				return
			}
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		claims := &jwt.RegisteredClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Key), nil
		})
		if required && err != nil {
			if cfg.UrlAuthRedirect != "" {
				c.Redirect(http.StatusFound, cfg.UrlAuthRedirect)
				return
			}
			c.AbortWithStatus(http.StatusForbidden)
			return
		} else if err == nil {
			c.Set(cfg.MidwStorage, claims.Subject)
		}
		c.Next()
		return
	}
}
func (cfg JWTConfig) JWTToken(data any) (string, error) {
	dta, err := json.Marshal(data)
	if err != nil {
		log.Err(err).Msg("Error marshalling data")
		return "", err
	}
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(cfg.ExpiresIn))),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    cfg.Issuer,
		Subject:   string(dta),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString([]byte(cfg.Key))

}
