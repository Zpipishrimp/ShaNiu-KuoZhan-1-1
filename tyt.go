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
					return "滚"
				}
				crons, err := qinglong.GetCrons("")
				if err != nil {
					return err
				}
				for _, cron := range crons {
					if strings.Contains(cron.Name, "推一推") {
						if cron.Pid != nil && fmt.Sprint(cron.Pid) != "" {
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
						if err := qinglong.Req(qinglong.CRONS, qinglong.PUT, "/run", []byte(fmt.Sprintf(`["%s"]`, cron.ID))); err != nil {
							return err
						}
						return "推起来啦。"
					}
				}
				return "推不动了。"
			},
		},
	})
}
