# pmsg

`pmsg` 是一个轻量级 Go 消息推送库，统一封装了邮件、钉钉自定义机器人和飞书自定义机器人。

## 功能

- 邮件：纯文本、HTML、抄送/密送、自定义邮件头和附件
- 钉钉：文本、链接、Markdown、ActionCard、FeedCard
- 飞书：文本、富文本、图片、群名片和消息卡片
- 钉钉、飞书机器人签名

## 环境要求

- Go 1.24.12 或更高版本
- 对应平台的 SMTP 账号或机器人 Webhook 凭证

## 安装

```bash
go get github.com/krabt/pmsg@latest
```

## 快速开始

### 钉钉

```go
package main

import (
	"log"

	pmsg "github.com/krabt/pmsg"
	"github.com/krabt/pmsg/bot/dingtalk"
)

func main() {
	client := pmsg.NewDingTalk("your-access-token", "your-secret")
	message := &dingtalk.Message{
		MsgType: dingtalk.MsgTypeMarkdown,
		Markdown: &dingtalk.MarkdownMeta{
			Title: "发布通知",
			Text:  "## 发布成功\n\n应用已更新。",
		},
	}

	if err := client.Send(message); err != nil {
		log.Fatal(err)
	}
}
```

未启用加签时可将 `secret` 传入空字符串。`accessToken` 只需填写 Webhook 中的令牌值。

### 飞书

```go
package main

import (
	"log"

	pmsg "github.com/krabt/pmsg"
	"github.com/krabt/pmsg/bot/feishu"
)

func main() {
	client := pmsg.NewFeishu("your-access-token", "your-secret")
	message := &feishu.Message{
		MsgType: feishu.MsgTypeText,
		Content: &feishu.ContentMeta{
			Text: "服务发布成功",
		},
	}

	if err := client.Send(message); err != nil {
		log.Fatal(err)
	}
}
```

未启用签名校验时可将 `secret` 传入空字符串。`accessToken` 只需填写 Webhook 路径末尾的令牌值。

### 邮件

```go
package main

import (
	"log"

	pmsg "github.com/krabt/pmsg"
	emailbot "github.com/krabt/pmsg/bot/email"
)

func main() {
	client := pmsg.NewEmail(
		"smtp.example.com",
		587,
		"username",
		"password",
		"sender@example.com",
	)
	client.StartTLS = true

	message := &emailbot.Message{
		To:      []string{"receiver@example.com"},
		Subject: "发布通知",
		Text:    "服务发布成功",
		HTML:    "<h1>发布成功</h1><p>服务已更新。</p>",
		Attachments: []emailbot.Attachment{
			{Path: "./report.pdf"},
		},
	}

	if err := client.Send(message); err != nil {
		log.Fatal(err)
	}
}
```

SMTP 加密方式：

- `client.TLS = true`：使用隐式 TLS，通常对应 465 端口。
- `client.StartTLS = true`：先建立普通 SMTP 连接，再升级到 TLS，通常对应 587 端口。
- 两者不能同时启用。
- 端口传 `0` 时默认使用 465 端口，并启用隐式 TLS。

附件既可以从文件读取，也可以直接传入内存数据：

```go
emailbot.Attachment{
	Filename:    "result.txt",
	ContentType: "text/plain",
	Data:        []byte("done"),
}
```

## 本地测试

普通单元测试和真实平台发送测试共用以下命令：

```bash
go test ./...
```

真实发送测试默认关闭。先复制配置模板：

```bash
cp .env.example .env
```

填写目标平台的凭证，并只打开需要测试的平台开关：

```dotenv
PMSG_TEST_DINGTALK=true
DINGTALK_ACCESS_TOKEN=your-access-token
DINGTALK_SECRET=your-secret
```

支持的开关如下：

| 开关 | 所需配置 |
| --- | --- |
| `PMSG_TEST_EMAIL` | `SMTP_HOST`、`SMTP_FROM`、`SMTP_TO`，以及服务器要求的认证信息 |
| `PMSG_TEST_DINGTALK` | `DINGTALK_ACCESS_TOKEN`，可选 `DINGTALK_SECRET` |
| `PMSG_TEST_FEISHU` | `FEISHU_ACCESS_TOKEN`，可选 `FEISHU_SECRET` |

可用 `PMSG_TEST_MESSAGE` 自定义测试消息。若配置文件不在项目根目录，可指定路径：

```bash
PMSG_ENV_FILE=/path/to/pmsg.env go test ./...
```

请勿提交 `.env`；仓库已将它加入 `.gitignore`。开启发送测试会向真实联系人或群聊发送消息，也可能受到平台频率限制。

## API 入口

| 平台 | 客户端构造函数 | 消息包 |
| --- | --- | --- |
| 邮件 | `pmsg.NewEmail(...)` | `github.com/krabt/pmsg/bot/email` |
| 钉钉 | `pmsg.NewDingTalk(...)` | `github.com/krabt/pmsg/bot/dingtalk` |
| 飞书 | `pmsg.NewFeishu(...)` | `github.com/krabt/pmsg/bot/feishu` |

所有平台的 `Send` 方法都会返回 `error`。网络错误、参数错误或平台返回非成功状态时，应记录或向上返回该错误。
