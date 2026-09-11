package dingtalk

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateMsgType(t *testing.T) {
	for _, value := range []string{MsgTypeText, MsgTypeLink, MsgTypeMarkdown, MsgTypeActionCard, MsgTypeSingleActionCard, MsgTypeFeedCard} {
		if err := ValidateMsgType(value); err != nil {
			t.Errorf("ValidateMsgType(%q) returned error: %v", value, err)
		}
	}
	if err := ValidateMsgType("unknown"); err == nil {
		t.Fatal("ValidateMsgType accepted an unknown type")
	}
}

func TestMessageJSON(t *testing.T) {
	data, err := json.Marshal(Message{MsgType: MsgTypeText, Text: &TextMeta{Content: "hello"}})
	if err != nil {
		t.Fatal(err)
	}
	encoded := string(data)
	if !strings.Contains(encoded, `"msgtype":"text"`) || !strings.Contains(encoded, `"text":{"content":"hello"}`) {
		t.Fatalf("message JSON = %s", encoded)
	}
	if strings.Contains(encoded, "markdown") {
		t.Fatalf("message JSON contains an unset field: %s", encoded)
	}
}
