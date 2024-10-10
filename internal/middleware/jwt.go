package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	jwtMiddleware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/vuisme/litecart/internal/models"
	"github.com/vuisme/litecart/internal/queries"
	"github.com/vuisme/litecart/pkg/webutil"
)

// JWTProtected is ...
func JWTProtected() func(*fiber.Ctx) error {
	config := jwtMiddleware.Config{
		KeyFunc:      customKeyFunc(),
		ContextKey:   "jwt",
		ErrorHandler: jwtError,
		TokenLookup:  "cookie:token",
	}

	return jwtMiddleware.New(config)
}

func jwtError(c *fiber.Ctx, err error) error {
	path := strings.Split(c.Path(), "/")[1]
	if path == "api" {
		if err.Error() == "Missing or malformed token" {
			return webutil.Response(c, fiber.StatusBadRequest, "Bad request", err.Error())
		}
		return webutil.Response(c, fiber.StatusUnauthorized, "Unauthorized", err.Error())
	}

	return c.Redirect("/_/signin")
}

func customKeyFunc() jwt.Keyfunc {
	return func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwtMiddleware.HS256 {
			return nil, fmt.Errorf("Unexpected jwt signing method=%v", t.Header["alg"])
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := queries.DB()
		settingJWT, _ := queries.GetSettingByGroup[models.JWT](ctx, db)
		return []byte(settingJWT.Secret), nil
	}
}
