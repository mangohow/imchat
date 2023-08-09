package resultcode

import "github.com/mangohow/easygin"

const (
	Unauthorized = iota + 1
	UpdateMessageFailed
	QueryFailed
)


var messager = map[int]string {
	UpdateMessageFailed: "更新消息状态失败",
	QueryFailed: "查询失败",
}



func init() {
	easygin.SetCodeMessager(func(i int) string {
		return messager[i]
	})
}