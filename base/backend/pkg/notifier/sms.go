package notifier

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

// SmsSender 短信发送器：对接一个「HTTP 网关」式的短信服务。
//
// 不同短信服务商的签名规则各不相同，底座不内置具体厂商 SDK，而是按统一约定调用网关：
//
//	POST {GatewayURL}
//	Authorization: Bearer {Token}（Token 为空则不发送该头）
//	{"to":"手机号","sign":"签名","subject":"标题","content":"内容"}
//
// 网关侧再转换为厂商 API（阿里云/腾讯云/梦网等），这样替换服务商只需改配置。
type SmsSender struct {
	GatewayURL string
	Token      string
	Sign       string
	client     *http.Client
}

func NewSmsSender(gatewayURL, token, sign string) *SmsSender {
	return &SmsSender{
		GatewayURL: gatewayURL,
		Token:      token,
		Sign:       sign,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SmsSender) Name() string {
	return "sms"
}

func (s *SmsSender) Send(msg Message) error {
	if s.GatewayURL == "" {
		return errors.New("未配置短信网关地址")
	}
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return errors.New("缺少接收手机号")
	}
	content := msg.Body
	if msg.Subject != "" {
		content = msg.Subject + " " + content
	}
	payload := map[string]string{
		"to":      to,
		"sign":    s.Sign,
		"subject": msg.Subject,
		"content": content,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.GatewayURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.Token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("短信网关返回状态码 " + resp.Status)
	}
	return nil
}
