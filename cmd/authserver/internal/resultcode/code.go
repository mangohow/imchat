package resultcode

const (
	UnknownError = -1

	PhoneInvalid = iota + 1
	PhoneCodeInvalid
	CodeGetFrequently
	PhoneHadBeenRegistered

	UsernameUnavailable
	UserInfoInvalid
	UsernameOrPasswordInvalid
	UserDestroyed
	UserBanned
	ParamInvalid

	QueryFailed
	OperationFailed
	ServerException
)


var messager = map[int]string{
	UnknownError: "未知错误，请重试",

	PhoneInvalid:      "请输入正确的手机号",
	PhoneCodeInvalid:  "手机验证码有误",
	CodeGetFrequently: "验证码获取频繁，请稍后再试",
	PhoneHadBeenRegistered: "该手机号已经被注册，不能重复注册",

	UsernameUnavailable: "用户名不可用",
	UserInfoInvalid: "用户信息格式非法",
	UsernameOrPasswordInvalid: "用户名或密码有误",
	UserDestroyed: "用户账号已注销",
	UserBanned: "用户已被封禁",
	ParamInvalid: "参数不合法",


	QueryFailed: "查询失败，请重试",
	OperationFailed: "操作失败，请重试",
	ServerException: "服务器异常",
}

func MessageFunc(code int) string {
	return messager[code]
}