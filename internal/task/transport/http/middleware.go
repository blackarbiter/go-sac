package http

import (
	"net/http"
	"strings"

	"github.com/blackarbiter/go-sac/pkg/utils/crypt"
	"github.com/gin-gonic/gin"
)

func JWTValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过Swagger文档路径
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		// 提取Token
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			return
		}

		// 验证Token
		claims, err := crypt.VerifyJWT(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// 注入用户ID到上下文
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
