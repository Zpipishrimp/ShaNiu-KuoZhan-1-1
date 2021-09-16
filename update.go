package jd_cookie

import (
	"fmt"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
	"github.com/cdle/sillyGirl/im"
)

func init() {
	core.AddCommand("", []core.Function{
		{
			Rules:   []string{`raw pt_key=([^;=\s]+);pt_pin=([^;=\s]+)`},
			Admin:   true,
			FindAll: true,
			Handle: func(s im.Sender) interface{} {
				value := fmt.Sprintf("pt_key=%s;pt_pin=%s;", s.Get(0), s.Get(1))
				envs, err := qinglong.GetEnvs(fmt.Sprintf(";pt_pin=%s;", s.Get(1)))
				if err != nil {
					return err
				}
				if len(envs) == 0 {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_COOKIE",
						Value: value,
					}); err != nil {
						return err
					}
					return "添加成功"
				} else {
					env := envs[0]
					env.Value = value
					if err := qinglong.UdpEnv(env); err != nil {
						return err
					}
					return "更新成功"
				}
			},
		},
		{
			Rules:   []string{`raw pin=([^;=\s]+);wskey=([^;=\s]+)`},
			Admin:   true,
			FindAll: true,
			Handle: func(s im.Sender) interface{} {
				value := fmt.Sprintf("pin=%s;wskey=%s;", s.Get(0), s.Get(1))
				envs, err := qinglong.GetEnvs(fmt.Sprintf("pin=%s;wskey=", s.Get(0)))
				if err != nil {
					return err
				}
				if len(envs) == 0 {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_WSCK",
						Value: value,
					}); err != nil {
						return err
					}
					return "添加成功"
				} else {
					env := envs[0]
					env.Value = value
					if err := qinglong.UdpEnv(env); err != nil {
						return err
					}
					return "更新成功"
				}
			},
		},
	})
}
