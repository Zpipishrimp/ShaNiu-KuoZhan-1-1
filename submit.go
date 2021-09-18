package jd_cookie

import (
	"fmt"
	"strings"
	"time"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
	"github.com/cdle/sillyGirl/im"
)

var pinQQ = core.NewBucket("pinQQ")
var pinTG = core.NewBucket("pinTG")

func init() {
	core.AddCommand("jd", []core.Function{
		{
			Rules: []string{`unbind ?`},
			Handle: func(s im.Sender) interface{} {
				s.Disappear(time.Second * 40)
				envs, err := qinglong.GetEnvs("JD_COOKIE")
				if err != nil {
					return err
				}
				if len(envs) == 0 {
					return "暂时无法操作。"
				}
				for _, env := range envs {
					pt_pin := FetchJdCookieValue("pt_pin", env.Value)
					pinQQ.Foreach(func(k, v []byte) error {
						if string(k) == pt_pin && string(v) == s.Get() {
							s.Reply(fmt.Sprintf("已解绑，%s。", pt_pin))
							defer func() {
								pinQQ.Set(string(k), "")
							}()
						}
						return nil
					})
					pinTG.Foreach(func(k, v []byte) error {
						if string(k) == pt_pin && string(v) == s.Get() {
							s.Reply(fmt.Sprintf("已解绑，%s。", pt_pin))
							defer func() {
								pinTG.Set(string(k), "")
							}()
						}
						return nil
					})
				}
				return "操作完成"
			},
		},
		{
			Rules:   []string{`raw pt_key=([^;=\s]+);\s*pt_pin=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s im.Sender) interface{} {
				s.Reply(s.Delete())
				s.Disappear(time.Second * 20)
				ck := &JdCookie{
					PtKey: s.Get(0),
					PtPin: s.Get(1),
				}
				if !ck.Available() {
					return "无效的ck，请重试。"
				}
				value := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
				envs, err := qinglong.GetEnvs(fmt.Sprintf("pt_pin=%s;", ck.PtPin))
				if err != nil {
					return err
				}
				if s.GetImType() == "qq" {
					pinQQ.Set(ck.PtPin, s.GetUserID())
				}
				if s.GetImType() == "tg" {
					pinTG.Set(ck.PtPin, s.GetUserID())
				}
				if len(envs) == 0 {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_COOKIE",
						Value: value,
					}); err != nil {
						return err
					}
					return ck.Nickname + ",添加成功。"
				} else {
					env := envs[0]
					env.Value = value
					if env.Status != 0 {
						if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err != nil {
							return err
						}
					}
					env.Status = 0
					if err := qinglong.UdpEnv(env); err != nil {
						return err
					}
					return ck.Nickname + ",更新成功。"
				}
			},
		},
		{
			Rules:   []string{`raw pin=([^;=\s]+);\s*wskey=([^;=\s]+)`},
			FindAll: true,
			Handle: func(s im.Sender) interface{} {
				s.Reply(s.Delete())
				s.Disappear(time.Second * 20)
				value := fmt.Sprintf("pin=%s;wskey=%s;", s.Get(0), s.Get(1))
				pt_key, err := getKey(value)
				if err == nil {
					if strings.Contains(pt_key, "fake") {
						return "无效的wskey，请重试。"
					}
				} else {
					s.Reply(err)
				}
				ck := &JdCookie{
					PtKey: pt_key,
					PtPin: s.Get(0),
				}
				ck.Available()
				envs, err := qinglong.GetEnvs(fmt.Sprintf("pin=%s;", ck.PtPin))
				if err != nil {
					return err
				}
				if s.GetImType() == "qq" {
					pinQQ.Set(ck.PtPin, s.GetUserID())
				}
				if s.GetImType() == "tg" {
					pinTG.Set(ck.PtPin, s.GetUserID())
				}
				var envCK *qinglong.Env
				var envWsCK *qinglong.Env
				for i := range envs {
					if strings.Contains(envs[i].Value, fmt.Sprintf("pin=%s;wskey=", ck.PtPin)) && envs[i].Name == "JD_WSCK" {
						envWsCK = &envs[i]
					} else if strings.Contains(envs[i].Value, fmt.Sprintf("pt_pin=%s;", ck.PtPin)) && envs[i].Name == "JD_COOKIE" {
						envCK = &envs[i]
					}
				}
				value2 := fmt.Sprintf("pt_key=%s;pt_pin=%s;", ck.PtKey, ck.PtPin)
				if envCK == nil {
					qinglong.AddEnv(qinglong.Env{
						Name:  "JD_COOKIE",
						Value: value2,
					})
				} else {
					envCK.Value = value2
					if err := qinglong.UdpEnv(*envCK); err != nil {
						return err
					}
				}
				if envWsCK == nil {
					if err := qinglong.AddEnv(qinglong.Env{
						Name:  "JD_WSCK",
						Value: value,
					}); err != nil {
						return err
					}
					return ck.Nickname + ",添加成功。"
				} else {
					env := envs[0]
					env.Value = value
					if env.Status != 0 {
						if err := qinglong.Req(qinglong.PUT, qinglong.ENVS, "/enable", []byte(`["`+env.ID+`"]`)); err != nil {
							return err
						}
					}
					env.Status = 0
					if err := qinglong.UdpEnv(env); err != nil {
						return err
					}
					return ck.Nickname + ",更新成功。"
				}
			},
		},
	})
}
