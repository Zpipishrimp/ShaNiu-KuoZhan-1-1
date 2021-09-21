package jd_cookie

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

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

type Session string

func (sess *Session) create() error {
	var address = jd_cookie.Get("address")
	if address == "" {
		return errors.New("未配置服务器地址")
	}
	html, _ := httplib.Get(address).String()
	res := regexp.MustCompile(`value="([\d\w]+)"`).FindStringSubmatch(html)
	if len(res) == 0 {
		return errors.New("匹配不到session")
	}
	var value = Session(res[1])
	sess = &value
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

func (sess *Session) sendAuthCode(phone string) error {
	address := jd_cookie.Get("address")
	req := httplib.Get(address + "sendAuthCode?clientSessionId=" + sess.String())
	_, err := req.Response()
	return err
}

func (sess *Session) String() string {
	return string(*sess)
}

func (sess *Session) query() (*Query, error) {
	query := &Query{}
	address := jd_cookie.Get("address")
	data, err := httplib.Get(fmt.Sprintf("%s/getScreen?clientSessionId=%s", address, sess.String())).Bytes()
	if err != nil {
		return nil, err
	}
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
		query, _ := sess.query()
		if query.PageStatus == "NORMAL" {
			break
		}
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

func (sess *Session) Polling() error {
	for {
		query, _ := sess.query()
		if query.PageStatus == "NORMAL" {
			break
		}
		if query.PageStatus == "SESSION_EXPIRED" {
			return errors.New("对不起，浏览器sessionId失效，请重新获取。")
		}
		if query.PageStatus == "VERIFY_FAILED_MAX" {
			return errors.New("验证码错误次数过多，请重新获取。")
		}
		if query.PageStatus == "VERIFY_CODE_MAX" {
			return errors.New("对不起，短信验证码发送次数已达上限，请24小时后再试。")
		}
		if query.PageStatus == "REQUIRE_VERIFY" {
			//正在破解滑块验证
		}
	}
	return nil
}
