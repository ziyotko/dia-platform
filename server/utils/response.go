package utils

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data,omitempty"`
}

func Success(msg string, data interface{}) *Response {
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
