package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func JWTMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token de autorización requerido",
			})
			c.Abort()
			return
		}

		// Extraer token del header "Bearer <token>"
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Verificar token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token inválido",
			})
			c.Abort()
			return
		}

		// Extraer claims y guardar en context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if uidRaw, ok := claims["user_id"].(string); ok {
				uid, err := uuid.Parse(uidRaw)
				if err == nil {
					c.Set("user_id", uid)
				}
			}
			c.Set("email", claims["email"])
		}

		c.Next()
	}
}
