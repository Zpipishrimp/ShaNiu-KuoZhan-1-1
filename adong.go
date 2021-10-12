package jd_cookie

import (
	"github.com/cdle/sillyGirl/core"
	"github.com/gin-gonic/gin"
)

func init() {
	core.Server.GET("/adong", func(c *gin.Context) {
		core.Senders <- &core.Faker{
			Message: c.Query("ck"),
			UserID:  c.Query("qq"),
			Type:    "qq",
		}
	})
}
