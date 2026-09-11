package wecom

import (
	"fmt"
	"net/url"

	"github.com/krabt/pmsg/utils"
)

const (
	MsgTypeText         = "text"          // 文本
	MsgTypeMarkdown     = "markdown"      // Markdown
	MsgTypeMarkdownV2   = "markdown_v2"   // Markdown V2
	MsgTypeImage        = "image"         // 图片
	MsgTypeNews         = "news"          // 图文
	MsgTypeFile         = "file"          // 文件
	MsgTypeVoice        = "voice"         // 语音
	MsgTypeTemplateCard = "template_card" // 模板卡片
)

// ValidateMsgType 验证消息类型
func ValidateMsgType(v string) error {
	switch v {
	case MsgTypeText,
		MsgTypeMarkdown,
		MsgTypeMarkdownV2,
		MsgTypeImage,
		MsgTypeNews,
		MsgTypeFile,
		MsgTypeVoice,
		MsgTypeTemplateCard:
	default:
		return fmt.Errorf(
			"%s not in [%q %q %q %q %q %q %q %q]",
			v,
			MsgTypeText,
			MsgTypeMarkdown,
			MsgTypeMarkdownV2,
			MsgTypeImage,
			MsgTypeNews,
			MsgTypeFile,
			MsgTypeVoice,
			MsgTypeTemplateCard,
		)
	}
	return nil
}

type Wecom struct {
	Key string
}

func New(key string) *Wecom {
	return &Wecom{
		Key: key,
	}
}

const sendURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="

// Send 发送企业微信群机器人消息
func (w *Wecom) Send(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	if err := ValidateMsgType(msg.MsgType); err != nil {
		return err
	}

	u := sendURL + url.QueryEscape(w.Key)

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
