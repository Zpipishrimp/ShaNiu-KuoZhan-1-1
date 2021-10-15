package jd_cookie

import (
	"github.com/cdle/sillyGirl/core"
	"github.com/gin-gonic/gin"
)

func init() {
	core.Server.Any("/adong", func(c *gin.Context) {
		core.Senders <- &core.Faker{
			Message: c.PostForm("ck"),
			UserID:  c.PostForm("qq"),
			Type:    "qq",
		}
	})
}
