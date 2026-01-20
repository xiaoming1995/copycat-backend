package middleware

import (
	"strings"

	"copycat/config"
	"copycat/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// 验证 Bearer Token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 解析 JWT Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.AppCfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// 提取用户 ID
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "invalid token claims")
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			response.Unauthorized(c, "invalid user id in token")
			c.Abort()
			return
		}

		// 将用户 ID 存入上下文
		c.Set("userID", int64(userID))
		c.Next()
	}
}
