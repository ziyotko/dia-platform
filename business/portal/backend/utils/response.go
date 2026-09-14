package utils

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

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

// PageData 列表统一返回体，固定为 {list,total,page,pageSize}，避免各控制器自行拼 gin.H 导致字段不一致。
// pageSize<1 表示不分页（全量返回），此时以实际返回条数回填，避免前端读到无意义的 0。
func PageData(list any, total int64, page, pageSize int) gin.H {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = sliceLen(list)
	}
	return gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}
}

// AllData 不分页的全量列表统一返回体（page=1，pageSize=列表长度）。
func AllData(list any, total int64) gin.H {
	return PageData(list, total, 1, 0)
}

// sliceLen 返回切片长度，非切片返回 0。
func sliceLen(list any) int {
	v := reflect.ValueOf(list)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len()
	}
	return 0
}
