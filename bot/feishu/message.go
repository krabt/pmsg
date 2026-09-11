package feishu

import (
	"fmt"
	"strconv"
	"time"

	"github.com/krabt/pmsg/utils"
)

const (
	MsgTypeText        = "text"        // 文本
	MsgTypePost        = "post"        // 富文本
	MsgTypeImage       = "image"       // 图片
	MsgTypeShareChat   = "share_chat"  // 分享群名片
	MsgTypeInteractive = "interactive" // 消息卡片
)

type Feishu struct {
	AccessToken string
	Secret      string
}

func New(accessToken string, secret string) *Feishu {
	return &Feishu{AccessToken: accessToken, Secret: secret}
}

// ValidateMsgType 验证
func ValidateMsgType(v string) error {
	switch v {
	case MsgTypeText, MsgTypePost, MsgTypeImage, MsgTypeShareChat, MsgTypeInteractive:
	default:
		return fmt.Errorf("%s not in [%q %q %q %q %q]", v,
			MsgTypeText, MsgTypePost, MsgTypeImage, MsgTypeShareChat, MsgTypeInteractive)
	}
	return nil
}

// Message 飞书自定义机器人消息
type Message struct {
	MsgType   string       `json:"msg_type"`            // 消息类型
	TimeStamp string       `json:"timestamp,omitempty"` // 为距当前时间不超过 1 小时(3600)的时间戳，时间单位s
	Sign      string       `json:"sign,omitempty"`      // 签名
	Content   *ContentMeta `json:"content,omitempty"`   // 消息内容
	Card      *CardMeta    `json:"card,omitempty"`      // 消息卡片
}

const sendURL = "https://open.feishu.cn/open-apis/bot/v2/hook/"

// Send 发送飞书自定义机器人消息
//
// 消息发送频率限制
// 自定义机器人的频率控制和普通应用不同，为单租户单机器人 100 次/分钟，5 次/秒。
// 建议发送消息尽量避开诸如 10:00、17:30 等整点及半点时间，否则可能出现因系统压力导致的 11232 限流错误，导致消息发送失败。
// 发送消息时，请求体的数据大小不能超过 20 KB。
func (f *Feishu) Send(msg *Message) error {
	if f.Secret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		msg.TimeStamp = timestamp
		msg.Sign = Sign(timestamp, f.Secret)
	}
	u := sendURL + f.AccessToken
	var resp ResponseMeta
	headers, err := utils.PostJSON(u, msg, &resp)
	if err != nil {
		if headers == nil {
			return err
		}
		return fmt.Errorf("%w, %s", err, resp.String())
	}
	if !resp.Succeed() {
		return fmt.Errorf("%s", resp.String())
	}
	return nil
}
