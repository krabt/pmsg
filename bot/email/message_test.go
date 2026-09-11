package email

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrepareBuildsAlternativeMessageAndHidesBcc(t *testing.T) {
	from, recipients, payload, err := prepare(&Email{
		Host: "smtp.example.com",
		Port: 587,
		From: "Alerts <alerts@example.com>",
	}, &Message{
		To:      []string{"Ops <ops@example.com>"},
		Cc:      []string{"cc@example.com"},
		Bcc:     []string{"hidden@example.com"},
		Subject: "服务告警",
		Text:    "plain",
		HTML:    "<p>html</p>",
		Attachments: []Attachment{
			{Filename: "details.json", ContentType: "application/json", Data: []byte(`{"status":"alert"}`)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if from != "alerts@example.com" {
		t.Fatalf("sender = %q", from)
	}
	if got, want := strings.Join(recipients, ","), "ops@example.com,cc@example.com,hidden@example.com"; got != want {
		t.Fatalf("recipients = %q, want %q", got, want)
	}
	encoded := string(payload)
	for _, value := range []string{"multipart/mixed", "multipart/alternative", "text/plain; charset=UTF-8", "text/html; charset=UTF-8", "details.json", "Content-Transfer-Encoding: base64", "eyJzdGF0dXMiOiJhbGVydCJ9"} {
		if !strings.Contains(encoded, value) {
			t.Errorf("message does not contain %q:\n%s", value, encoded)
		}
	}
	if strings.Contains(encoded, "hidden@example.com") {
		t.Errorf("Bcc address leaked into message headers:\n%s", encoded)
	}
}

func TestAttachmentDataReadsPathAndValidatesFilename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("attachment body"), 0o600); err != nil {
		t.Fatal(err)
	}
	filename, data, err := attachmentData(Attachment{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	if filename != "message.txt" || string(data) != "attachment body" {
		t.Fatalf("attachmentData = (%q, %q)", filename, data)
	}
	if _, _, err := attachmentData(Attachment{Data: []byte("data")}); err == nil {
		t.Fatal("attachmentData succeeded without a filename")
	}
}

func TestPrepareRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config *Email
		msg    *Message
	}{
		{"missing host", &Email{Port: 25, From: "from@example.com"}, &Message{To: []string{"to@example.com"}}},
		{"invalid port", &Email{Host: "smtp.example.com", Port: -1, From: "from@example.com"}, &Message{To: []string{"to@example.com"}}},
		{"no recipients", &Email{Host: "smtp.example.com", Port: 25, From: "from@example.com"}, &Message{}},
		{"both TLS modes", &Email{Host: "smtp.example.com", Port: 25, From: "from@example.com", TLS: true, StartTLS: true}, &Message{To: []string{"to@example.com"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, _, err := prepare(test.config, test.msg); err == nil {
				t.Fatal("prepare succeeded, want error")
			}
		})
	}
}

func TestApplyDefaultsUsesSMTPS(t *testing.T) {
	config := applyDefaults(&Email{Host: "smtp.example.com"})
	if config.Port != 465 || !config.TLS {
		t.Fatalf("defaults = port %d, TLS %t; want port 465 with TLS", config.Port, config.TLS)
	}
}

func TestEncodeMessageRejectsManagedHeaderOverride(t *testing.T) {
	_, err := encodeMessage("from@example.com", &Message{
		To:      []string{"to@example.com"},
		Headers: map[string]string{"from": "other@example.com"},
	})
	if err == nil {
		t.Fatal("encodeMessage succeeded, want error")
	}
}
