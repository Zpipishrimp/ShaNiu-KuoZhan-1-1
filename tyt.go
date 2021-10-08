package jd_cookie

import (
	"fmt"
	"strings"

	"github.com/cdle/sillyGirl/core"
	"github.com/cdle/sillyGirl/develop/qinglong"
)

func init() {
	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw packetId=(\S+)(&|&amp;)currentActId`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if s.GetImType() == "tg" {
					return nil //文明用语
				}
				crons, err := qinglong.GetCrons("")
				if err != nil {
					return err
				}
				for _, cron := range crons {
					if strings.Contains(cron.Name, "推一推") {
						if cron.Status == 0 { //修复错误
							return "推一推已在运行中。"
						}
						err := qinglong.SetConfigEnv(qinglong.Env{
							Name:   "tytpacketId",
							Value:  s.Get(),
							Status: 3,
						})
						if err != nil {
							return err
						}
						if err := qinglong.Config.Req(qinglong.CRONS, qinglong.PUT, "/run", []byte(fmt.Sprintf(`["%s"]`, cron.ID))); err != nil {
							return err
						}
						return "推一推起来啦。"
					}
				}
				return "推一推不知道为啥推不动了。"
			},
		},
	})
}
