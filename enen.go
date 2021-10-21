package jd_cookie

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/adapter/httplib"
	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
)

func init() {
	core.AddCommand("jd", []core.Function{
		{
			Rules: []string{"enen ?"},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				req := httplib.Get("https://plogin.m.jd.com/cgi-bin/ml/mlogout?appid=300&returnurl=https%3A%2F%2Fm.jd.com%2F")
				req.Header("authority", "plogin.m.jd.com")
				req.Header("User-Agent", ua())
				req.Header("cookie", s.Get())
				req.Header("host", "jd.com")
				req.Response()
				return "已注销登录"
			},
		},
		{
			Rules: []string{"eueu ?"},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				envs, err := qinglong.GetEnvs("JD_WSCK")
				if err != nil {
					return err
				}
				yes := false
				for _, env := range envs {
					if strings.Contains(env.Value, s.Get()) {
						yes = true
						pin := core.FetchCookieValue("pin", env.Value)
						pt_key, err := getKey(env.Value)
						if err != nil {
							return err
						}
						s.Reply(fmt.Sprintf("pt_key=%s;pt_pin=%s;", pt_key, pin))
					}
				}
				if !yes {
					return "找不到转换目标"
				}
				return nil
			},
		},
	})

}
