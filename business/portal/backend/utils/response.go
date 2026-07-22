package utils

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data any    `json:"data,omitempty"`
}

func Success(msg string, data any) *Response {
	if msg == "" {
		msg = "success"
	}
	return &Response{
		Code: 0,
		Msg:  msg,
		Data: data,
	}
}

func Error(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}
