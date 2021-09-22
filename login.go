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

func (sess *Session) create() error {
	var address = jd_cookie.Get("address")
	if address == "" {
		return errors.New("未配置服务器地址")
	}
	html, _ := httplib.Get(address).String()
	res := regexp.MustCompile(`value="([\d\w]+)"`).FindStringSubmatch(html)
	if len(res) == 0 {
		return errors.New(jd_cookie.Get("login_fail", "崩了请找作者，仓库地址：https://github.com/rubyangxg/jd-qinglong"))
	}
	sess.Value = res[1]
	return nil
}

func (sess *Session) control(name, value string) error {
	address := jd_cookie.Get("address")
	req := httplib.Post(address + "/control")
	req.Param("currId", name)
	req.Param("currValue", value)
	req.Param("clientSessionId", sess.String())
	_, err := req.String()
	// fmt.Println("controll", name, value, rt)
	return err
}

func (sess *Session) login(phone, sms_code string) error {
	address := jd_cookie.Get("address")
	req := httplib.Post(address + "/jdLogin")
	req.Param("phone", phone)
	req.Param("sms_code", sms_code)
	req.Param("clientSessionId", sess.String())
	_, err := req.String()
	// fmt.Println(phone, sms_code, rt)
	return err
}

func (sess *Session) sendAuthCode() error {
	address := jd_cookie.Get("address")
	req := httplib.Get(address + "/sendAuthCode?clientSessionId=" + sess.String())
	_, err := req.Response()
	return err
}

func (sess *Session) String() string {
	return sess.Value
}

func (sess *Session) query() (*Query, error) {
	query := &Query{}
	address := jd_cookie.Get("address")
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
	address := jd_cookie.Get("address")
	_, err := httplib.Get(fmt.Sprintf("%s/crackCaptcha?clientSessionId=%s", address, sess.String())).Response()
	return err
}

var codes map[string]chan string

func init() {
	codes = map[string]chan string{}
	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw ^(\d{11})$`},
			Handle: func(s core.Sender) interface{} {
				if jd_cookie.Get("igtg", false) == "true" && s.GetImType() == "tg" {
					return "滚，不欢迎你。"
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
				defer delete(codes, id)
				var sess = new(Session)
				phone := s.Get()
				if err := sess.Phone(phone); err != nil {
					return err
				}
				send := false
				login := false
				verify := false
				sms_code := ""
				for {
					query, _ := sess.query()
					if query.PageStatus == "SESSION_EXPIRED" {
						return errors.New("登录超时。")
					}
					if query.SessionTimeOut == 0 {
						return errors.New("登录超时。")
					}
					if query.CanClickLogin && !login {
						s.Reply("正在登录...")
						if err := sess.login(phone, sms_code); err != nil {
							return err
						}
					}
					if query.PageStatus == "VERIFY_FAILED_MAX" {
						return errors.New("验证码错误次数过多，请重新获取。")
					}
					if query.PageStatus == "VERIFY_CODE_MAX" {
						return errors.New("对不起，短信验证码请求频繁，请稍后再试。")
					}
					if query.PageStatus == "REQUIRE_VERIFY" && !verify {
						verify = true
						s.Reply("正在自动验证...")
						if err := sess.crackCaptcha(); err != nil {
							return err
						}
						s.Reply("验证通过。")
						s.Reply("请输入验证码______")
						select {
						case sms_code = <-c:
							s.Reply("正在提交验证码...")
							if err := sess.SmsCode(sms_code); err != nil {
								return err
							}
							s.Reply("验证码提交成功。")
						case <-time.After(60 * time.Second):
							return "验证码超时。"
						}
					}
					if query.CanSendAuth && !send {
						if err := sess.sendAuthCode(); err != nil {
							return err
						}
						send = true
					}
					if !query.CanSendAuth && query.AuthCodeCountDown > 0 {

					}
					if query.AuthCodeCountDown == -1 && send {
						// return "验证码超时。"
					}
					if query.PageStatus == "SUCCESS_CK" {
						core.Senders <- &core.Faker{
							Message: fmt.Sprintf("pt_key=%v;pt_pin=%v;", query.Ck.PtKey, query.Ck.PtPin),
							UserID:  s.GetUserID(),
							Type:    s.GetImType(),
						}
						s.Reply("登录成功！")
						return nil
					}
					time.Sleep(time.Second)
				}
			},
		},
		{
			Rules: []string{`raw ^登录$`},
			Handle: func(s core.Sender) interface{} {
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
				return "你要登上敌方的陆地？"
			},
		},

		{
			Rules: []string{`raw ^(\d{6})$`},
			Handle: func(s core.Sender) interface{} {
				if code, ok := codes[s.GetImType()+fmt.Sprint(s.GetUserID())]; ok {
					code <- s.Get()
				} else {
					s.Reply("验证码不存在。")
				}
				return nil
			},
		},
	})
}
