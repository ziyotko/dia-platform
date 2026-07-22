package notifier

import "errors"

// Message 通用消息
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender 消息发送器接口
type Sender interface {
	Send(msg Message) error
	Name() string
}

var senders = make(map[string]Sender)

func Register(name string, sender Sender) {
	senders[name] = sender
}

func Get(name string) (Sender, error) {
	sender, ok := senders[name]
	if !ok {
		return nil, errors.New("未找到发送器: " + name)
	}
	return sender, nil
}

func Send(channel string, msg Message) error {
	sender, err := Get(channel)
	if err != nil {
		return err
	}
	return sender.Send(msg)
}
