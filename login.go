package jd_cookie

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/beego/beego/v2/adapter/httplib"
	"github.com/cdle/sillyGirl/core"
)

var jd_cookie = core.NewBucket("jd_cookie")

type Query struct {
	// Screen interface{} `json:"screen"`
	Ck struct {
		PtPin interface{} `json:"ptPin"`
		PtKey interface{} `json:"ptKey"`
		Empty bool        `json:"empty"`
	} `json:"ck"`
	PageStatus        string `json:"pageStatus"`
	AuthCodeCountDown int    `json:"authCodeCountDown"`
	CanClickLogin     bool   `json:"canClickLogin"`
	CanSendAuth       bool   `json:"canSendAuth"`
	SessionTimeOut    int    `json:"sessionTimeOut"`
	AvailChrome       int    `json:"availChrome"`
}

type Session struct {
	Value string
}

func (sess *Session) address() string {
	if v := regexp.MustCompile(`^(https?://[\.\w]+:?\d*)`).FindStringSubmatch(jd_cookie.Get("address")); len(v) == 2 {
		return v[1]
	}
	return ""
}

func (sess *Session) create() error {
	var address = sess.address()
	var url = "https://github.com/rubyangxg/jd-qinglong"
	if address == "" {
		return errors.New("未配置服务器地址，仓库地址：" + url)
	}
	req := httplib.Get(address)
	req.SetTimeout(time.Second, time.Second)
	html, err := req.String()
	if err != nil {
		return err
	}
	res := regexp.MustCompile(`value="([\d\w]+)"`).FindStringSubmatch(html)
	if len(res) == 0 {
		return errors.New(jd_cookie.Get("login_fail", "崩了请找作者，仓库地址：https://github.com/rubyangxg/jd-qinglong"+url))
	}
	sess.Value = res[1]
	return nil
}

func (sess *Session) control(name, value string) error {
	address := sess.address()
	req := httplib.Post(address + "/control")
	req.Param("currId", name)
	req.Param("currValue", value)
	req.Param("clientSessionId", sess.String())
	_, err := req.String()
	// fmt.Println("controll", name, value, rt)
	return err
}

func (sess *Session) login(phone, sms_code string) error {
	address := sess.address()
	req := httplib.Post(address + "/jdLogin")
	req.Param("phone", phone)
	req.Param("sms_code", sms_code)
	req.Param("clientSessionId", sess.String())
	_, err := req.String()
	// fmt.Println(phone, sms_code, rt)
	return err
}

func (sess *Session) sendAuthCode() error {
	address := sess.address()
	req := httplib.Get(address + "/sendAuthCode?clientSessionId=" + sess.String())
	_, err := req.Response()
	return err
}

func (sess *Session) String() string {
	return sess.Value
}

func (sess *Session) query() (*Query, error) {
	query := &Query{}
	address := sess.address()
	// fmt.Println(sess.String(), "+++")
	data, err := httplib.Get(fmt.Sprintf("%s/getScreen?clientSessionId=%s", address, sess.String())).Bytes()
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(data))
	err = json.Unmarshal(data, &query)
	if err != nil {
		return nil, err
	}
	return query, nil
}

func (sess *Session) Phone(phone string) error {
	err := sess.create()
	if err != nil {
		return err
	}
	for {
		query, err := sess.query()
		if err != nil {
			return err
		}
		if query.PageStatus == "NORMAL" {
			break
		}
		if query.PageStatus == "SESSION_EXPIRED" {
			return sess.Phone(phone)
		}
		time.Sleep(time.Second)
	}
	err = sess.control("phone", phone)
	if err != nil {
		return err
	}
	return nil
}

func (sess *Session) SmsCode(sms_code string) error {
	err := sess.control("sms_code", sms_code)
	if err != nil {
		return err
	}
	return nil
}

func (sess *Session) crackCaptcha() error {
	address := sess.address()
	req := httplib.Get(fmt.Sprintf("%s/crackCaptcha?clientSessionId=%s", address, sess.String()))
	req.SetTimeout(time.Second*10, time.Second*10)
	_, err := req.Response()
	return err
}

func (sess *Session) releaseSession() error {
	address := sess.address()
	req := httplib.Get(fmt.Sprintf("%s/releaseSession?clientSessionId=%s", address, sess.String()))
	_, err := req.Response()
	return err
}

var codes map[string]chan string

func init() {
	codes = map[string]chan string{}
	core.BeforeStop = append(core.BeforeStop, func() {
		for {
			if len(codes) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
	})
	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw ^(\d{11})$`},
			Handle: func(s core.Sender) interface{} {
				s.Delete()
				if groupCode := jd_cookie.GetInt("groupCode"); !s.IsAdmin() && groupCode != 0 && s.GetChatID() != 0 && groupCode != s.GetChatID() {
					s.Reply("傻妞已崩溃。")
					return nil
				}
				if num := jd_cookie.GetInt("login_num", 2); len(codes) >= num {
					return fmt.Sprintf("%v坑位全部在使用中，请排队。", num)
				}
				id := s.GetImType() + fmt.Sprint(s.GetUserID())
				if _, ok := codes[id]; ok {
					return "你已在登录中。"
				}
				c := make(chan string, 1)
				codes[id] = c
				var sess = new(Session)
				phone := s.Get()
				err := sess.create()
				if err != nil {
					delete(codes, id)
					return err
				}
				go func() {
					defer delete(codes, id)
					defer sess.releaseSession()
					s.Reply("请稍后，正在模拟环境...", core.E)
					for {
						query, err := sess.query()
						if err != nil {
							s.Reply(err, core.E)
							return
						}
						if query.PageStatus == "NORMAL" {
							break
						}
						if query.PageStatus == "SESSION_EXPIRED" {
							sess.create()
						}
						time.Sleep(time.Second)
					}
					err = sess.control("phone", phone)
					if err != nil {
						s.Reply(err, core.E)
						return
					}
					send := false
					login := false
					verify := false
					success := false
					sms_code := ""
					for {
						query, err := sess.query()
						if err != nil {
							s.Reply(err, core.E)
							return
						}
						if query.PageStatus == "SESSION_EXPIRED" {
							if !login {
								s.Reply(errors.New("登录超时。"), core.E)
							}
							return
						}
						if query.SessionTimeOut == 0 {
							if success {
								return
							}
							s.Reply(errors.New("登录超时。"), core.E)
							return
						}
						if query.CanClickLogin && !login {
							s.Reply("正在登录...", core.E)
							if err := sess.login(phone, sms_code); err != nil {
								s.Reply(err, core.E)
								return
							}
							login = true
						}
						if query.PageStatus == "REQUIRE_VERIFY" && !verify {
							verify = true
							s.Reply("正在自动验证...", core.E)
							if err := sess.crackCaptcha(); err != nil {
								s.Reply(err, core.E)
								return
							}
							s.Reply("验证通过。", core.E)
							s.Reply("请输入验证码______", core.E)
							timeout := 1
							for {
								select {
								case sms_code = <-c:
									s.Reply("正在提交验证码...", core.E)
									if err := sess.SmsCode(sms_code); err != nil {
										s.Reply(err, core.E)
										return
									}
									s.Reply("验证码提交成功。", core.E)
									goto HELL
								case <-time.After(time.Millisecond * 300):
									query, err := sess.query()
									if err != nil {
										s.Reply(err, core.E)
										return
									}
									if query.PageStatus == "SESSION_EXPIRED" {
										goto HELL
									}
									if query.PageStatus == "VERIFY_FAILED_MAX" {
										s.Reply("验证码错误次数过多，请重新获取。", core.E)
										return
									}
									if query.PageStatus == "VERIFY_CODE_MAX" || query.PageStatus == "SWITCH_SMS_LOGIN" {
										s.Reply("对不起，短信验证码请求频繁，请稍后再试。", core.E)
										return
									}
									if query.AuthCodeCountDown <= 0 {
										timeout++
										if timeout > 20 {
											s.Reply("验证码超时，登录失败。", core.E)
											return
										}
									}
								}
							}
						HELL:
						}
						if query.CanSendAuth && !send {
							if err := sess.sendAuthCode(); err != nil {
								s.Reply(err, core.E)
								return
							}
							send = true
						}
						if !query.CanSendAuth && query.AuthCodeCountDown > 0 {

						}
						if query.AuthCodeCountDown == -1 && send {

						}
						if query.PageStatus == "SUCCESS_CK" && !success {
							cookie := fmt.Sprintf("pt_key=%v;pt_pin=%v;", query.Ck.PtKey, query.Ck.PtPin)
							qq := ""
							if s.GetImType() == "qq" {
								qq = fmt.Sprint(s.GetUserID())
							}
							xdd(cookie, qq)
							core.Senders <- &core.Faker{
								Message: cookie,
								UserID:  s.GetUserID(),
								Type:    s.GetImType(),
							}
							s.Reply("登录成功，你可以登录下一个账号。", core.E)
							success = true
							return
						}
						time.Sleep(time.Second)
					}
				}()
				if s.GetImType() == "wxmp" {
					return "一会儿收到验证码发给我哦～"
				}
				return nil
			},
		},
		{
			Rules: []string{`raw ^登录$`},
			Handle: func(s core.Sender) interface{} {
				if groupCode := jd_cookie.GetInt("groupCode"); !s.IsAdmin() && groupCode != 0 && s.GetChatID() != 0 && groupCode != s.GetChatID() {
					s.Delete()
					s.Disappear()
					s.Reply("我崩溃了。")
					return nil
				}
				if num := jd_cookie.GetInt("login_num", 2); len(codes) >= num {
					return fmt.Sprintf("%v坑位全部在使用中，请排队(稍后再试)。", num)
				}
				id := s.GetImType() + fmt.Sprint(s.GetUserID())
				if _, ok := codes[id]; ok {
					return "你已在登录中。"
				}
				s.Reply("请输入手机号___________")
				return nil
			},
		},
		{
			Rules: []string{`raw ^登陆$`},
			Handle: func(s core.Sender) interface{} {
				if num := jd_cookie.GetInt("login_num", 2); len(codes) >= num {
					return fmt.Sprintf("%v坑位全部在使用中，请排队(稍后再试)。", num)
				}
				id := s.GetImType() + fmt.Sprint(s.GetUserID())
				if _, ok := codes[id]; ok {
					return "你已在登录中。"
				}
				s.Reply("你要登上敌方的陆地？")
				s.Reply("请输入手机号___________", time.Duration(time.Second*5))
				return nil
			},
		},

		{
			Rules: []string{`raw ^(\d{6})$`},
			Handle: func(s core.Sender) interface{} {
				s.Delete()
				if code, ok := codes[s.GetImType()+fmt.Sprint(s.GetUserID())]; ok {
					code <- s.Get()
					if s.GetImType() == "wxmp" {
						s.Reply("八九不离十登录成功了，一分钟后对我说\"查询\"确认是否登录成功。")
					}
				} else {
					s.Reply("验证码不存在或过期了。")
				}
				return nil
			},
		},
	})
}
