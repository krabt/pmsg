package wecom

// Message 企业微信群机器人消息
type Message struct {
	MsgType      string          `json:"msgtype"`
	Text         *TextMeta       `json:"text,omitempty"`
	Markdown     *MarkdownMeta   `json:"markdown,omitempty"`
	MarkdownV2   *MarkdownV2Meta `json:"markdown_v2,omitempty"`
	Image        *ImageMeta      `json:"image,omitempty"`
	News         *NewsMeta       `json:"news,omitempty"`
	File         *FileMeta       `json:"file,omitempty"`
	Voice        *VoiceMeta      `json:"voice,omitempty"`
	TemplateCard any             `json:"template_card,omitempty"`
}

// TextMeta 文本消息
type TextMeta struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// MarkdownMeta Markdown消息
type MarkdownMeta struct {
	Content string `json:"content"`
}

// MarkdownV2Meta Markdown V2消息
type MarkdownV2Meta struct {
	Content string `json:"content"`
}

// ImageMeta 图片消息
type ImageMeta struct {
	Base64 string `json:"base64"`
	MD5    string `json:"md5"`
}

// NewsMeta 图文消息
type NewsMeta struct {
	Articles []ArticleMeta `json:"articles"`
}

// ArticleMeta 图文
type ArticleMeta struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl,omitempty"`
}

// FileMeta 文件消息
type FileMeta struct {
	MediaID string `json:"media_id"`
}

// VoiceMeta 语音消息
type VoiceMeta struct {
	MediaID string `json:"media_id"`
}
