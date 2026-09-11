package pmsg2

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/krabt/pmsg/bot/dingtalk"
	emailbot "github.com/krabt/pmsg/bot/email"
	"github.com/krabt/pmsg/bot/feishu"
	"github.com/krabt/pmsg/bot/wecom"
)

func TestNewEmail(t *testing.T) {
	got := NewEmail("smtp.example.com", 587, "user", "pass", "from@example.com")
	if got.Host != "smtp.example.com" || got.Port != 587 || got.Username != "user" || got.Password != "pass" || got.From != "from@example.com" {
		t.Fatalf("NewEmail() = %+v", got)
	}
}

func TestNewBots(t *testing.T) {
	ding := NewDingTalk("token", "secret")
	if ding.AccessToken != "token" || ding.Secret != "secret" {
		t.Fatalf("NewDingTalk() = %+v", ding)
	}

	wecom := NewWecom("key")
	if wecom.Key != "key" {
		t.Fatalf("NewWecom() = %+v", wecom)
	}
}

func TestSendEmailFromDotEnv(t *testing.T) {
	env := sendTestEnv(t, "PMSG_TEST_EMAIL")
	port, err := strconv.Atoi(envValue(env, "SMTP_PORT", "587"))
	if err != nil {
		t.Fatalf("invalid SMTP_PORT: %v", err)
	}
	recipients := strings.Split(env["SMTP_TO"], ",")
	if env["SMTP_HOST"] == "" || env["SMTP_FROM"] == "" || env["SMTP_TO"] == "" {
		t.Fatal("SMTP_HOST, SMTP_FROM and SMTP_TO are required")
	}
	msg := &emailbot.Message{To: recipients, Subject: "pmsg test", Text: envValue(env, "PMSG_TEST_MESSAGE", "pmsg test message")}
	config := NewEmail(env["SMTP_HOST"], port, env["SMTP_USERNAME"], env["SMTP_PASSWORD"], env["SMTP_FROM"])
	config.TLS = envBool(env, "SMTP_TLS")
	config.StartTLS = envBool(env, "SMTP_STARTTLS")
	if err := config.Send(msg); err != nil {
		t.Fatal(err)
	}
}

func TestSendDingTalkFromDotEnv(t *testing.T) {
	env := sendTestEnv(t, "PMSG_TEST_DINGTALK")
	token := requiredEnv(t, env, "DINGTALK_ACCESS_TOKEN")
	message := envValue(env, "PMSG_TEST_MESSAGE", "pmsg test message")
	messages := []*dingtalk.Message{
		{MsgType: dingtalk.MsgTypeText, Text: &dingtalk.TextMeta{Content: message}},
		{MsgType: dingtalk.MsgTypeLink, Link: &dingtalk.LinkMeta{Title: "pmsg test", Text: message, MessageUrl: "https://example.com"}},
		{MsgType: dingtalk.MsgTypeMarkdown, Markdown: &dingtalk.MarkdownMeta{Title: "pmsg test", Text: message}},
		{MsgType: dingtalk.MsgTypeActionCard, ActionCard: &dingtalk.ActionCardMeta{Title: "pmsg test", Text: message, Btns: []dingtalk.ActionCardBtnMeta{{Title: "Open", ActionURL: "https://example.com"}}}},
		{MsgType: dingtalk.MsgTypeFeedCard, FeedCard: &dingtalk.FeedCardMeta{Links: []dingtalk.FeedCardLinkMeta{{Title: "pmsg test", MessageURL: "https://example.com", PicURL: "https://example.com/image.png"}}}},
	}
	for _, msg := range messages {
		if err := NewDingTalk(token, env["DINGTALK_SECRET"]).Send(msg); err != nil {
			t.Errorf("send DingTalk %s message: %v", msg.MsgType, err)
		}
	}
}

func TestSendFeishuFromDotEnv(t *testing.T) {
	env := sendTestEnv(t, "PMSG_TEST_FEISHU")
	token := requiredEnv(t, env, "FEISHU_ACCESS_TOKEN")
	client := feishu.New(token, env["FEISHU_SECRET"])
	// imageKey := requiredEnv(t, env, "FEISHU_IMAGE_KEY")
	// shareChatID := requiredEnv(t, env, "FEISHU_SHARE_CHAT_ID")
	for _, msg := range feishuTestMessages(envValue(env, "PMSG_TEST_MESSAGE", "pmsg test message"), "", "") {
		if err := client.Send(msg); err != nil {
			t.Errorf("send Feishu %s message: %v", msg.MsgType, err)
		}
	}
}

func TestSendWecomFromDotEnv(t *testing.T) {
	env := sendTestEnv(t, "PMSG_TEST_WECOM")
	key := requiredEnv(t, env, "WECOM_KEY")
	message := envValue(env, "PMSG_TEST_MESSAGE", "pmsg test message")
	messages := map[string]*wecom.Message{
		wecom.MsgTypeText:       wecom.NewText(message),
		wecom.MsgTypeMarkdown:   wecom.NewMarkdown(message),
		wecom.MsgTypeMarkdownV2: wecom.NewMarkdownV2(message),
		// wecom.MsgTypeImage:        wecom.NewImage([]byte(message)),
		wecom.MsgTypeNews:         {MsgType: wecom.MsgTypeNews, News: &wecom.NewsMeta{Articles: []wecom.ArticleMeta{{Title: "pmsg test", Description: message, URL: "https://example.com"}}}},
		wecom.MsgTypeTemplateCard: {MsgType: wecom.MsgTypeTemplateCard, TemplateCard: map[string]any{"card_type": "text_notice", "main_title": map[string]any{"title": "pmsg test", "desc": message}, "card_action": map[string]any{"type": 1, "url": "https://example.com"}}},
	}
	if mediaID := env["WECOM_MEDIA_ID"]; mediaID != "" {
		messages[wecom.MsgTypeFile] = wecom.NewFile(mediaID)
		messages[wecom.MsgTypeVoice] = &wecom.Message{MsgType: wecom.MsgTypeVoice, Voice: &wecom.VoiceMeta{MediaID: mediaID}}
	}
	client := NewWecom(key)
	for msgType, msg := range messages {
		t.Run(msgType, func(t *testing.T) {
			if err := client.Send(msg); err != nil {
				t.Fatalf("send WeCom %s message: %v", msg.MsgType, err)
			}
		})
	}
	for _, msgType := range []string{wecom.MsgTypeFile, wecom.MsgTypeVoice} {
		if _, ok := messages[msgType]; !ok {
			t.Run(msgType, func(t *testing.T) {
				t.Skip("set WECOM_MEDIA_ID in .env to test this message type")
			})
		}
	}
}

func feishuTestMessages(message, imageKey, shareChatID string) []*feishu.Message {
	messages := []*feishu.Message{
		{MsgType: feishu.MsgTypeText, Content: &feishu.ContentMeta{Text: message}},
		{MsgType: feishu.MsgTypePost, Content: &feishu.ContentMeta{Post: &feishu.PostMeta{ZhCn: feishu.PostZhCn{
			Title:   "pmsg test",
			Content: [][]feishu.PostZhCnContent{{{Tag: "text", Text: message}}},
		}}}},
		// {MsgType: feishu.MsgTypeImage, Content: &feishu.ContentMeta{ImageKey: imageKey}},
		// {MsgType: feishu.MsgTypeShareChat, Content: &feishu.ContentMeta{ShareChatID: shareChatID}},
		{MsgType: feishu.MsgTypeInteractive, Card: &feishu.CardMeta{
			Header:   feishu.CardHeader{Title: feishu.CardHeaderTitle{Content: "pmsg test", Tag: "plain_text"}},
			Elements: []feishu.CardElement{{Tag: "div", Text: feishu.CardElementText{Content: message, Tag: "plain_text"}}},
		}},
	}
	return messages
}

func sendTestEnv(t *testing.T, control string) map[string]string {
	t.Helper()
	env, err := readDotEnv(dotEnvPath())
	if errors.Is(err, os.ErrNotExist) {
		t.Skip(".env not found")
	}
	if err != nil {
		t.Fatal(err)
	}
	if !envBool(env, control) {
		t.Skip("set " + control + "=true in .env to enable this send test")
	}
	return env
}

func requiredEnv(t *testing.T, env map[string]string, key string) string {
	t.Helper()
	if value := env[key]; value != "" {
		return value
	}
	t.Fatalf("%s is required", key)
	return ""
}

func dotEnvPath() string {
	if path := os.Getenv("PMSG_ENV_FILE"); path != "" {
		return path
	}
	return ".env"
}

func readDotEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func envValue(values map[string]string, key, fallback string) string {
	if value := values[key]; value != "" {
		return value
	}
	return fallback
}

func envBool(values map[string]string, key string) bool {
	value, err := strconv.ParseBool(values[key])
	return err == nil && value
}
