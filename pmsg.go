package pmsg2

import (
	"github.com/krabt/pmsg/bot/dingtalk"
	"github.com/krabt/pmsg/bot/email"
	"github.com/krabt/pmsg/bot/feishu"
	"github.com/krabt/pmsg/bot/wecom"
)

func NewEmail(Host string, Port int, Username string, Password string, From string) *email.Email {
	return email.New(Host, Port, Username, Password, From)
}

func NewDingTalk(accessToken string, secret string) *dingtalk.Dingtalk {
	return dingtalk.New(accessToken, secret)
}

func NewFeishu(accessToken string, secret string) *feishu.Feishu {
	return feishu.New(accessToken, secret)
}

func NewWecom(key string) *wecom.Wecom {
	return wecom.New(key)
}
