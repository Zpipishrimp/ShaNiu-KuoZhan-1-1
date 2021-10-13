package jd_cookie

import (
	"fmt"
	"strings"

	"github.com/cdle/sillyGirl/core"
)

var jd_cookie = core.NewBucket("jd_cookie")

func init() {

	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw ^登录$`, `raw ^登陆$`},
			Handle: func(s core.Sender) interface{} {
				if groupCode := jd_cookie.Get("groupCode"); !s.IsAdmin() && groupCode != "" && s.GetChatID() != 0 && !strings.Contains(groupCode, fmt.Sprint(s.GetChatID())) {
					return nil
				}
				// 	if c == nil {
				tip := jd_cookie.Get("tip")
				if tip == "" {
					if s.IsAdmin() {
						return jd_cookie.Get("tip", "阿东不行啦，更改登录提示指令，set jd_cookie tip ?")
					} else {
						tip = "暂时无法使用短信登录。"
					}
				}

				return tip
				// 	}
				// 	uid := s.GetUserID()
				// 	stop := false
				// 	go func() {
				// 		for {

				// 		}
				// 	}()
				// 	for {
				// 		if stop == true {
				// 			break
				// 		}
				// 		s.Await(s, func(s core.Sender) interface{} {
				// 			msg := s.GetContent()
				// 			if strings.Contains(msg, "退出") {
				// 				stop = true
				// 				return nil
				// 			}
				// 			c.WriteJSON(map[string]interface{}{
				// 				"time":         time.Now().Unix(),
				// 				"self_id":      uid,
				// 				"post_type":    "message",
				// 				"message_type": "private",
				// 				"sub_type":     "friend",
				// 				"message_id":   s.GetMessageID(),
				// 				"user_id":      12345678,
				// 				"message":      s.GetContent(),
				// 				"raw_message":  s.GetContent(),
				// 				"font":         456,
				// 				"sender": map[string]interface{}{
				// 					"nickname": "傻妞",
				// 					"sex":      "female",
				// 					"age":      18,
				// 				},
				// 			})
				// 			return nil
				// 		}, `[\s\S]+`)
				// 	}
				// 	return "已退出登录模式"
			},
		},
	})
}

// var c *websocket.Conn

// func RunServer() {
// 	addr := jd_cookie.Get("adong_addr")
// 	if addr == "" {
// 		return
// 	}
// 	defer func() {
// 		time.Sleep(time.Second * 2)
// 		RunServer()
// 	}()
// 	u := url.URL{Scheme: "ws", Host: addr, Path: "/wx/event"}
// 	logs.Info("连接阿东 %s", u.String())
// 	var err error
// 	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
// 	if err != nil {
// 		logs.Warn("连接阿东错误:", err)
// 	}
// 	defer c.Close()
// 	go func() {
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				log.Println("read:", err)
// 				return
// 			}
// 			log.Printf("recv: %s", message)
// 		}
// 	}()
// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case t := <-ticker.C:
// 			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
// 			if err != nil {
// 				log.Println("阿东错误:", err)
// 				c = nil
// 				return
// 			}
// 		}
// 	}
// }
