package wecom

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
)

func NewText(content string) *Message {
	return &Message{
		MsgType: MsgTypeText,
		Text: &TextMeta{
			Content: content,
		},
	}
}

func NewTextAt(content string, users ...string) *Message {
	return &Message{
		MsgType: MsgTypeText,
		Text: &TextMeta{
			Content:       content,
			MentionedList: users,
		},
	}
}

func NewTextAtAll(content string) *Message {
	return NewTextAt(content, "@all")
}

func NewMarkdown(content string) *Message {
	return &Message{
		MsgType: MsgTypeMarkdown,
		Markdown: &MarkdownMeta{
			Content: content,
		},
	}
}

func NewMarkdownV2(content string) *Message {
	return &Message{
		MsgType: MsgTypeMarkdownV2,
		MarkdownV2: &MarkdownV2Meta{
			Content: content,
		},
	}
}

func NewImage(data []byte) *Message {
	sum := md5.Sum(data)

	return &Message{
		MsgType: MsgTypeImage,
		Image: &ImageMeta{
			Base64: base64.StdEncoding.EncodeToString(data),
			MD5:    hex.EncodeToString(sum[:]),
		},
	}
}

func NewFile(mediaID string) *Message {
	return &Message{
		MsgType: MsgTypeFile,
		File: &FileMeta{
			MediaID: mediaID,
		},
	}
}
