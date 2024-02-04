package filter

import (
	"net/http"

	"github.com/apisix/manager-api/internal/conf"
	"github.com/gin-gonic/gin"
)

func Ldap() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/apisix/admin/ldap/login" {
			c.Next()
			return
		}

		if c.Request.URL.Path == "/apisix/admin/ldap/logout" {
			cookie, _ := conf.CookieStore.Get(c.Request, "oidc")
			if cookie.IsNew {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			cookie.Options.MaxAge = -1
			cookie.Save(c.Request, c.Writer)
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}

}
