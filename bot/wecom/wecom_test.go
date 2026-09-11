package wecom

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMessageConstructors(t *testing.T) {
	if got := NewText("hello"); got.MsgType != MsgTypeText || got.Text.Content != "hello" {
		t.Fatalf("NewText() = %+v", got)
	}
	if got := NewTextAt("hello", "user1", "user2"); strings.Join(got.Text.MentionedList, ",") != "user1,user2" {
		t.Fatalf("NewTextAt() = %+v", got)
	}
	if got := NewTextAtAll("hello"); len(got.Text.MentionedList) != 1 || got.Text.MentionedList[0] != "@all" {
		t.Fatalf("NewTextAtAll() = %+v", got)
	}
	if got := NewMarkdown("# title"); got.MsgType != MsgTypeMarkdown || got.Markdown.Content != "# title" {
		t.Fatalf("NewMarkdown() = %+v", got)
	}
	if got := NewMarkdownV2("# title"); got.MsgType != MsgTypeMarkdownV2 || got.MarkdownV2.Content != "# title" {
		t.Fatalf("NewMarkdownV2() = %+v", got)
	}
	if got := NewFile("media-id"); got.MsgType != MsgTypeFile || got.File.MediaID != "media-id" {
		t.Fatalf("NewFile() = %+v", got)
	}
}

func TestNewImage(t *testing.T) {
	got := NewImage([]byte("image"))
	if got.MsgType != MsgTypeImage || got.Image.Base64 != "aW1hZ2U=" || got.Image.MD5 != "78805a221a988e79ef3f42d7c5bfd418" {
		t.Fatalf("NewImage() = %+v", got)
	}
}

func TestValidateMsgType(t *testing.T) {
	valid := []string{MsgTypeText, MsgTypeMarkdown, MsgTypeMarkdownV2, MsgTypeImage, MsgTypeNews, MsgTypeFile, MsgTypeVoice, MsgTypeTemplateCard}
	for _, value := range valid {
		if err := ValidateMsgType(value); err != nil {
			t.Errorf("ValidateMsgType(%q) returned error: %v", value, err)
		}
	}
	if err := ValidateMsgType("unknown"); err == nil {
		t.Fatal("ValidateMsgType accepted an unknown type")
	}
}

func TestResponseMeta(t *testing.T) {
	if !(&ResponseMeta{}).Succeed() {
		t.Fatal("zero ResponseMeta should succeed")
	}
	got := (&ResponseMeta{ErrCode: 40001, ErrMsg: "invalid"}).String()
	if got != "errcode=40001, errmsg=invalid" {
		t.Fatalf("String() = %q", got)
	}
}

func TestMessageJSONOmitsUnsetFields(t *testing.T) {
	data, err := json.Marshal(NewText("hello"))
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, `"msgtype":"text"`) || !strings.Contains(encoded, `"content":"hello"`) {
		t.Fatalf("message JSON = %s", encoded)
	}
	if strings.Contains(encoded, "markdown") || strings.Contains(encoded, "image") {
		t.Fatalf("message JSON contains unset fields: %s", encoded)
	}
}
