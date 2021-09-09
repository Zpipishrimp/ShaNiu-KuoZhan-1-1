package jd_cookie

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/jd_asset"
	"github.com/cdle/sillyGirl/develop/qinglong"
	"github.com/cdle/sillyGirl/im"
)

func init() {
	core.AddCommand("jd", []core.Function{
		{
			Rules: []string{`find ?`},
			Handle: func(s im.Sender) interface{} {
				a := s.Get()
				envs, err := qinglong.GetEnvs("JD_COOKIE")
				if err != nil {
					return err
				}
				if len(envs) == 0 {
					return "青龙未设置京东账号。"
				}
				ncks := []qinglong.Env{}
				if s := strings.Split(a, "-"); len(s) == 2 {
					for i := range envs {
						if i+1 >= jd_asset.Int(s[0]) && i+1 <= jd_asset.Int(s[1]) {
							ncks = append(ncks, envs[i])
						}
					}
				} else if x := regexp.MustCompile(`^[\s\d,]+$`).FindString(a); x != "" {
					xx := regexp.MustCompile(`(\d+)`).FindAllStringSubmatch(a, -1)
					for i := range envs {
						for _, x := range xx {
							if fmt.Sprint(i+1) == x[1] {
								ncks = append(ncks, envs[i])
							}
						}

					}
				} else if a != "" {
					a = strings.Replace(a, " ", "", -1)
					for i := range envs {
						if strings.Contains(envs[i].Value, a) || strings.Contains(envs[i].Remarks, a) {
							ncks = append(ncks, envs[i])
						}
					}
				}
				if len(ncks) == 0 {
					return "没有匹配的京东账号。"
				}
				msg := []string{}
				for _, ck := range ncks {
					status := "已启用"
					if ck.Status != 0 {
						status = "已禁用"
					}
					msg = append(msg, strings.Join([]string{
						fmt.Sprintf("编号：%v", ck.ID),
						fmt.Sprintf("状态：%v", status),
						fmt.Sprintf("值：%v", ck.Value),
					}, "\n"))
				}
				return strings.Join(msg, "\n\n")
			},
		},
	})
}
