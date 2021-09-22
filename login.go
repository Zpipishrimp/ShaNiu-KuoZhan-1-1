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
	Screen interface{} `json:"screen"`
	Ck     struct {
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
		return errors.New("其他用户正在使用，请稍后再试。")
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
	_, err := req.Response()
	return err
}

func (sess *Session) login(phone, sms_code string) error {
	address := jd_cookie.Get("address")
	req := httplib.Post(address + "/control")
	req.Param("phone", phone)
	req.Param("sms_code", sms_code)
	req.Param("clientSessionId", sess.String())
	_, err := req.Response()
	return err
}

func (sess *Session) sendAuthCode() error {
	address := jd_cookie.Get("address")
	req := httplib.Get(address + "sendAuthCode?clientSessionId=" + sess.String())
	_, err := req.Response()
	return err
}

func (sess *Session) String() string {
	return sess.Value
}

func (sess *Session) query() (*Query, error) {
	query := &Query{}
	address := jd_cookie.Get("address")
	fmt.Println(sess.String(), "+++")
	data, err := httplib.Get(fmt.Sprintf("%s/getScreen?clientSessionId=%s", address, sess.String())).Bytes()
	if err != nil {
		return nil, err
	}
	fmt.Println(string(data))
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
	core.AddCommand("", []core.Function{
		{
			Rules: []string{`raw ^(\d{11})$`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if _, ok := codes[s.GetImType()+fmt.Sprint(s.GetUserID())]; ok {
					return "你已在登录中."
				}
				id := s.GetImType() + fmt.Sprint(s.GetUserID())
				defer delete(codes, id)
				var sess = new(Session)
				phone := s.Get()
				if err := sess.Phone(phone); err != nil {
					return err
				}
				for {
					query, _ := sess.query()
					if query.PageStatus == "NORMAL" {
						continue
					}
					if query.PageStatus == "SESSION_EXPIRED" {
						return errors.New("对不起，登录超时。")
					}
					if query.SessionTimeOut == 0 {
						return errors.New("对不起，登录超时。")
					}
					if query.CanClickLogin {
						//可以点击登录
						c := make(chan string, 1)

						codes[id] = c
						select {
						case sms_code := <-c:
							sess.login(phone, sms_code)
						case <-time.After(60 * time.Second):
							return "验证码超时。"
						}
					}
					if query.PageStatus == "VERIFY_FAILED_MAX" {
						return errors.New("验证码错误次数过多，请重新获取。")
					}
					if query.PageStatus == "VERIFY_CODE_MAX" {
						return errors.New("对不起，短信验证码请求频繁，请稍后再试。")
					}
					if query.PageStatus == "REQUIRE_VERIFY" {
						sess.crackCaptcha()
					}
					if query.CanSendAuth {
						sess.sendAuthCode()
						s.Reply("请输入验证码__")
					}
					if !query.CanSendAuth && query.AuthCodeCountDown > 0 {

					}
					if query.PageStatus == "SUCCESS_CK" {
						return fmt.Sprintf("pt_key=%v;pt_pin=%v;", query.Ck.PtKey, query.Ck.PtPin)
					}
					time.Sleep(time.Second)
				}
			},
		},
		{
			Rules: []string{`raw ^登录$`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if _, ok := codes[s.GetImType()+fmt.Sprint(s.GetUserID())]; ok {
					return "你已在登录中."
				}
				s.Reply("请输入手机号__")
				return nil
			},
		},
		{
			Rules: []string{`raw ^\d{5}$`},
			Admin: true,
			Handle: func(s core.Sender) interface{} {
				if code, ok := codes[s.GetImType()+fmt.Sprint(s.GetUserID())]; ok {
					code <- s.Get()
				} else {
					s.Reply("验证码已过期。")
				}
				return nil
			},
		},
	})
}
