package jd_cookie

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
)

func init() {
	core.AddCommand("jd", []core.Function{
		{
			Rules: []string{`find ?`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				a := s.Get()
				envs, err := qinglong.GetEnvs("JD_COOKIE")
				if err != nil {
					return err
				}
				if len(envs) == 0 {
					return "青龙未设置变量。"
				}
				ncks := []qinglong.Env{}
				if s := strings.Split(a, "-"); len(s) == 2 {
					for i := range envs {
						if i+1 >= core.Int(s[0]) && i+1 <= core.Int(s[1]) {
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
						if strings.Contains(envs[i].Value, a) || strings.Contains(envs[i].Remarks, a) || strings.Contains(envs[i].ID, a) {
							ncks = append(ncks, envs[i])
						}
					}
				}
				if len(ncks) == 0 {
					return "没有匹配的变量。"
				}
				msg := []string{}
				for _, ck := range ncks {
					status := "已启用"
					if ck.Status != 0 {
						status = "已禁用"
					}
					msg = append(msg, strings.Join([]string{
						fmt.Sprintf("编号：%v", ck.ID),
						fmt.Sprintf("备注：%v", ck.Remarks),
						fmt.Sprintf("状态：%v", status),
						fmt.Sprintf("pin值：%v", core.FetchCookieValue(ck.Value, "pt_pin")),
					}, "\n"))
				}
				return strings.Join(msg, "\n\n")
			},
		},
		{
			Rules: []string{`exchange ? ?`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				ac1 := s.Get(0)
				ac2 := s.Get(1)
				envs, err := qinglong.GetEnvs("JD_COOKIE")
				if err != nil {
					return err
				}
				if len(envs) < 2 {
					return "数目小于，无需交换顺序。"
				}
				toe := []qinglong.Env{}
				for _, env := range envs {
					if env.ID == ac1 || env.ID == ac2 {
						toe = append(toe, env)
					}
				}
				if len(toe) < 2 {
					return "找不到对应的变量，无法交换顺序。"
				}
				toe[0].ID, toe[1].ID = toe[1].ID, toe[0].ID
				if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, toe[0]); err != nil {
					return err
				}
				if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, toe[1]); err != nil {
					return err
				}
				return "交换成功"
			},
		},
		{
			Rules: []string{`enable ?`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+s.Get()+`"]`)); err != nil {
					return err
				}
				return "操作成功"
			},
		},
		{
			Rules: []string{`disable ?`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, "/disable", []byte(`["`+s.Get()+`"]`)); err != nil {
					return err
				}
				return "操作成功"
			},
		},
		{
			Rules: []string{`remark ? ?`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				env, err := qinglong.GetEnv(s.Get())
				if err != nil {
					return err
				}
				env.Remarks = s.Get(1)
				if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, *env); err != nil {
					return err
				}
				return "操作成功"
			},
		},
	})
}
