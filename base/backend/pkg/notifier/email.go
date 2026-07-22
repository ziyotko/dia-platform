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

func MustLoadEmailSender(get func(key string) string) {
	port, _ := strconv.Atoi(get("email_port"))
	if get("email_host") == "" || get("email_username") == "" {
		return
	}
	cfg := EmailConfig{
		Host:     get("email_host"),
		Port:     port,
		Username: get("email_username"),
		Password: get("email_password"),
		From:     get("email_from"),
		SSL:      get("email_ssl") == "true",
	}
	Register("email", NewEmailSender(cfg))
}
