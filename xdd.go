package jd_cookie

import "github.com/beego/beego/v2/client/httplib"

func xdd(cookie string, qq string) {
	xdd_url := jd_cookie.Get("xdd_url")
	if xdd_url != "" {
		req := httplib.Post(xdd_url)
		req.Param("ck", cookie)
		req.Param("qq", qq)
		req.Response()
	}
}
