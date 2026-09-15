package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// WeChatSender 企业微信群机器人 webhook 发送器。
// 群机器人是「一个 webhook 对应一个群」，因此整条消息只推送一次，不区分接收人。
type WeChatSender struct {
	WebhookURL string
	client     *http.Client
}

func NewWeChatSender(webhookURL string) *WeChatSender {
	return &WeChatSender{
		WebhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *WeChatSender) Name() string {
	return "wechat"
}

func (s *WeChatSender) Send(msg Message) error {
	if s.WebhookURL == "" {
		return errors.New("未配置企业微信机器人 webhook 地址")
	}
	text := msg.Subject
	if msg.Body != "" {
		if text != "" {
			text += "\n"
		}
		text += msg.Body
	}
	payload := map[string]interface{}{
		"msgtype": "text",
		"text":    map[string]string{"content": text},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := s.client.Post(s.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("企业微信 webhook 返回状态码 " + resp.Status)
	}
	return nil
}
