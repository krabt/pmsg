package dingtalk

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/krabt/pmsg/utils"
)

const (
	MsgTypeText             = "text"              // 文本
	MsgTypeLink             = "link"              // 链接
	MsgTypeMarkdown         = "markdown"          // markdown
	MsgTypeFeedCard         = "feedCard"          // FeedCard
	MsgTypeActionCard       = "actionCard"        // ActionCard
	MsgTypeSingleActionCard = "single_actionCard" // Single ActionCard
)

// ValidateMsgType 验证
func ValidateMsgType(v string) error {
	switch v {
	case MsgTypeText, MsgTypeLink, MsgTypeMarkdown, MsgTypeActionCard, MsgTypeSingleActionCard, MsgTypeFeedCard:
	default:
		return fmt.Errorf("%s not in [%q %q %q %q %q %q]", v,
			MsgTypeText, MsgTypeLink, MsgTypeMarkdown, MsgTypeActionCard, MsgTypeSingleActionCard, MsgTypeFeedCard)
	}
	return nil
}

type Dingtalk struct {
	AccessToken string
	Secret      string
}

func New(accessToken string, secret string) *Dingtalk {
	return &Dingtalk{
		AccessToken: accessToken,
		Secret:      secret,
	}
}

// Message 钉钉自定义机器人消息
type Message struct {
	MsgType    string        `json:"msgtype"`              // 消息类型
	Text       *TextMeta     `json:"text,omitempty"`       // 文本消息
	Markdown   *MarkdownMeta `json:"markdown,omitempty"`   // markdown消息
	At         *AtMeta       `json:"at,omitempty"`         // @
	Link       *LinkMeta     `json:"link,omitempty"`       // 链接
	ActionCard any           `json:"actionCard,omitempty"` // ActionCard
	FeedCard   *FeedCardMeta `json:"feedCard,omitempty"`   // FeedCard
}

const sendURL = "https://oapi.dingtalk.com/robot/send?access_token="

// Send 发送钉钉自定义机器人消息
//
// 消息发送频率限制
// 每个机器人每分钟最多发送20条消息到群里，如果超过20条，会限流10分钟
// 如果你有大量发消息的场景（譬如系统监控报警）可以将这些信息进行整合，通过markdown消息以摘要的形式发送到群里。
func (d *Dingtalk) Send(msg *Message) error {
	u := sendURL + url.QueryEscape(d.AccessToken)
	if d.Secret != "" {
		timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		sign, err := Sign(timestamp, d.Secret)
		if err != nil {
			return fmt.Errorf("sign failed: %w", err)
		}
		u = u + "&timestamp=" + timestamp + "&sign=" + url.QueryEscape(sign)
	}
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
