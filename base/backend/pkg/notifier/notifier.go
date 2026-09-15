package notifier

import (
	"errors"
	"sort"
)

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

// Names 返回当前已注册（已配置）的发送器名称，供前端渲染可选的发送渠道。
func Names() []string {
	names := make([]string, 0, len(senders))
	for name := range senders {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Unregister 注销发送器（配置被清空时使用，避免渠道仍可选）。
func Unregister(name string) {
	delete(senders, name)
}
