package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const dialTimeout = 10 * time.Second

// Config contains the SMTP server settings.
//
// When Port is zero, pmsg uses SMTPS on port 465 and enables TLS. TLS
// establishes an implicit TLS connection (usually port 465). StartTLS upgrades
// a normal SMTP connection using STARTTLS (usually port 587). They cannot be
// used together.
type Email struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      bool
	StartTLS bool
}

func New(Host string, Port int, Username string, Password string, From string) *Email {
	return &Email{
		Host:     Host,
		Port:     Port,
		Username: Username,
		Password: Password,
		From:     From,
	}
}

// Message is an email message. Text and HTML may be used independently or
// together; when both are set, recipients receive a multipart alternative.
// Bcc recipients receive the message but are deliberately omitted from its
// headers.
type Message struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Text        string
	HTML        string
	Attachments []Attachment
	Headers     map[string]string
}

// Attachment is a file included in a Message. Set Path to attach a local file,
// or Data to attach bytes already held in memory. Filename overrides the base
// name of Path and is required when using Data.
type Attachment struct {
	Filename    string
	ContentType string
	Path        string
	Data        []byte
}

// Send delivers msg using config's SMTP server.
func (e *Email) Send(msg *Message) error {
	from, recipients, payload, err := prepare(e, msg)
	if err != nil {
		return err
	}

	address := net.JoinHostPort(e.Host, strconv.Itoa(e.Port))
	conn, err := net.DialTimeout("tcp", address, dialTimeout)
	if err != nil {
		return fmt.Errorf("dial smtp server: %w", err)
	}

	if e.TLS {
		conn = tls.Client(conn, &tls.Config{ServerName: e.Host, MinVersion: tls.VersionTLS12})
	}
	client, err := smtp.NewClient(conn, e.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("create smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if e.StartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("smtp server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: e.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("start TLS: %w", err)
		}
	}
	if e.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", e.Username, e.Password, e.Host)); err != nil {
			return fmt.Errorf("smtp authentication: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("set sender: %w", err)
	}
	for _, recipient := range recipients {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("set recipient %q: %w", recipient, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start message data: %w", err)
	}
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write message data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish message data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit smtp session: %w", err)
	}
	return nil
}

func prepare(config *Email, msg *Message) (string, []string, []byte, error) {
	config = applyDefaults(config)
	if config.Host == "" {
		return "", nil, nil, errors.New("smtp host is required")
	}
	if config.Port < 1 || config.Port > 65535 {
		return "", nil, nil, fmt.Errorf("invalid smtp port %d", config.Port)
	}
	if config.TLS && config.StartTLS {
		return "", nil, nil, errors.New("TLS and STARTTLS cannot both be enabled")
	}
	if msg == nil {
		return "", nil, nil, errors.New("email message is required")
	}
	from, err := mailbox(config.From)
	if err != nil {
		return "", nil, nil, fmt.Errorf("invalid sender: %w", err)
	}
	recipientHeaders := append([]string{}, msg.To...)
	recipientHeaders = append(recipientHeaders, msg.Cc...)
	recipientHeaders = append(recipientHeaders, msg.Bcc...)
	recipients, err := mailboxes(recipientHeaders)
	if err != nil {
		return "", nil, nil, fmt.Errorf("invalid recipient: %w", err)
	}
	if len(recipients) == 0 {
		return "", nil, nil, errors.New("at least one recipient is required")
	}
	payload, err := encodeMessage(config.From, msg)
	if err != nil {
		return "", nil, nil, err
	}
	return from, recipients, payload, nil
}

func applyDefaults(config *Email) *Email {
	if config.Port == 0 {
		config.Port = 465
		config.TLS = true
	}
	return config
}

func mailbox(value string) (string, error) {
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address == "" {
		if err == nil {
			err = errors.New("empty address")
		}
		return "", err
	}
	return address.Address, nil
}

func mailboxes(values []string) ([]string, error) {
	result := make([]string, 0, len(values))
	for _, value := range values {
		address, err := mailbox(value)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", value, err)
		}
		result = append(result, address)
	}
	return result, nil
}

func encodeMessage(from string, msg *Message) ([]byte, error) {
	if containsNewline(msg.Subject) {
		return nil, errors.New("subject must not contain a newline")
	}
	var buf bytes.Buffer
	headers := textproto.MIMEHeader{}
	headers.Set("From", from)
	if len(msg.To) > 0 {
		headers.Set("To", strings.Join(msg.To, ", "))
	}
	if len(msg.Cc) > 0 {
		headers.Set("Cc", strings.Join(msg.Cc, ", "))
	}
	headers.Set("Subject", mime.QEncoding.Encode("UTF-8", msg.Subject))
	headers.Set("Date", time.Now().Format(time.RFC1123Z))
	headers.Set("MIME-Version", "1.0")
	for key, value := range msg.Headers {
		if key == "" || containsNewline(key) || containsNewline(value) {
			return nil, errors.New("email headers must not contain a newline")
		}
		if _, exists := headers[textproto.CanonicalMIMEHeaderKey(key)]; exists {
			return nil, fmt.Errorf("header %q is managed by email.Message", key)
		}
		headers.Set(key, value)
	}

	if len(msg.Attachments) == 0 {
		contentType, err := writeBody(&buf, msg)
		if err != nil {
			return nil, err
		}
		headers.Set("Content-Type", contentType)
		body := append([]byte(nil), buf.Bytes()...)
		buf.Reset()
		writeHeaders(&buf, headers)
		buf.WriteString("\r\n")
		buf.Write(body)
		return buf.Bytes(), nil
	}

	mixed := multipart.NewWriter(&buf)
	headers.Set("Content-Type", "multipart/mixed; boundary="+mixed.Boundary())
	writeHeaders(&buf, headers)
	buf.WriteString("\r\n")
	if err := writeBodyPart(mixed, msg); err != nil {
		return nil, err
	}
	for _, attachment := range msg.Attachments {
		if err := writeAttachment(mixed, attachment); err != nil {
			return nil, err
		}
	}
	if err := mixed.Close(); err != nil {
		return nil, fmt.Errorf("close multipart message: %w", err)
	}
	return buf.Bytes(), nil
}

func writeBody(buffer *bytes.Buffer, msg *Message) (string, error) {
	if msg.HTML == "" {
		buffer.WriteString(msg.Text)
		return "text/plain; charset=UTF-8", nil
	}
	if msg.Text == "" {
		buffer.WriteString(msg.HTML)
		return "text/html; charset=UTF-8", nil
	}
	alternative := multipart.NewWriter(buffer)
	for _, part := range []struct {
		contentType string
		body        string
	}{{"text/plain; charset=UTF-8", msg.Text}, {"text/html; charset=UTF-8", msg.HTML}} {
		writer, err := alternative.CreatePart(textproto.MIMEHeader{"Content-Type": {part.contentType}})
		if err != nil {
			return "", fmt.Errorf("create message part: %w", err)
		}
		if _, err := io.WriteString(writer, part.body); err != nil {
			return "", fmt.Errorf("write message part: %w", err)
		}
	}
	if err := alternative.Close(); err != nil {
		return "", fmt.Errorf("close multipart alternative: %w", err)
	}
	return "multipart/alternative; boundary=" + alternative.Boundary(), nil
}

func writeBodyPart(mixed *multipart.Writer, msg *Message) error {
	var body bytes.Buffer
	contentType, err := writeBody(&body, msg)
	if err != nil {
		return err
	}
	writer, err := mixed.CreatePart(textproto.MIMEHeader{"Content-Type": {contentType}})
	if err != nil {
		return fmt.Errorf("create message body: %w", err)
	}
	if _, err := writer.Write(body.Bytes()); err != nil {
		return fmt.Errorf("write message body: %w", err)
	}
	return nil
}

func writeAttachment(mixed *multipart.Writer, attachment Attachment) error {
	filename, data, err := attachmentData(attachment)
	if err != nil {
		return err
	}
	contentType := attachment.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	partHeaders := textproto.MIMEHeader{}
	partHeaders.Set("Content-Type", mime.FormatMediaType(contentType, map[string]string{"name": filename}))
	partHeaders.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	partHeaders.Set("Content-Transfer-Encoding", "base64")
	writer, err := mixed.CreatePart(partHeaders)
	if err != nil {
		return fmt.Errorf("create attachment %q: %w", filename, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	for len(encoded) > 0 {
		line := encoded
		if len(line) > 76 {
			line = line[:76]
		}
		if _, err := io.WriteString(writer, line+"\r\n"); err != nil {
			return fmt.Errorf("encode attachment %q: %w", filename, err)
		}
		encoded = encoded[len(line):]
	}
	return nil
}

func attachmentData(attachment Attachment) (string, []byte, error) {
	filename := attachment.Filename
	data := attachment.Data
	if attachment.Path != "" {
		fileData, err := os.ReadFile(attachment.Path)
		if err != nil {
			return "", nil, fmt.Errorf("read attachment %q: %w", attachment.Path, err)
		}
		data = fileData
		if filename == "" {
			filename = filepath.Base(attachment.Path)
		}
	}
	if filename == "" || filename == "." || containsNewline(filename) {
		return "", nil, errors.New("attachment filename is required and must not contain a newline")
	}
	return filename, data, nil
}

func writeHeaders(buf *bytes.Buffer, headers textproto.MIMEHeader) {
	for key, values := range headers {
		for _, value := range values {
			_, _ = fmt.Fprintf(buf, "%s: %s\r\n", key, value)
		}
	}
}

func containsNewline(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}
