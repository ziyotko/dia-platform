package notifier

import (
	"strconv"

	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	SSL      bool
}

type EmailSender struct {
	cfg EmailConfig
}

func NewEmailSender(cfg EmailConfig) *EmailSender {
	return &EmailSender{cfg: cfg}
}

func (s *EmailSender) Name() string {
	return "email"
}

func (s *EmailSender) Send(msg Message) error {
	m := gomail.NewMessage()
	from := s.cfg.From
	if from == "" {
		from = s.cfg.Username
	}
	m.SetHeader("From", from)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/html", msg.Body)

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.Username, s.cfg.Password)
	d.SSL = s.cfg.SSL
	return d.DialAndSend(m)
}

// LoadEmailSender 按「系统设置 → 邮件配置」（category=email）的键名加载邮件发送器。
// 注意键名不带前缀：host / port / username / password / from / ssl。
func LoadEmailSender(get func(key string) string) {
	port, _ := strconv.Atoi(get("port"))
	if get("host") == "" || get("username") == "" {
		Unregister("email")
		return
	}
	cfg := EmailConfig{
		Host:     get("host"),
		Port:     port,
		Username: get("username"),
		Password: get("password"),
		From:     get("from"),
		SSL:      get("ssl") == "true",
	}
	Register("email", NewEmailSender(cfg))
}

// LoadWeChatSender 加载企业微信群机器人发送器（未配置 webhook 时不注册该渠道）。
func LoadWeChatSender(get func(key string) string) {
	url := get("wechat_webhook_url")
	if url == "" {
		Unregister("wechat")
		return
	}
	Register("wechat", NewWeChatSender(url))
}

// LoadSmsSender 加载短信网关发送器（未配置网关地址时不注册该渠道）。
func LoadSmsSender(get func(key string) string) {
	url := get("sms_gateway_url")
	if url == "" {
		Unregister("sms")
		return
	}
	Register("sms", NewSmsSender(url, get("sms_gateway_token"), get("sms_sign")))
}
