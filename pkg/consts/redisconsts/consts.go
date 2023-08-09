package redisconsts

import "time"

const (
	LoginPhoneKey = "login:phoneCode:"
	TokenKey = "token:id:"
	UserKey = "user:id:"
	UserInfoKey = "userinfo:id:"
	FriendsKey = "friends:id:"
	FriendKey = "friend:id:"

	UserCounterKey = "user:counter"
)

const (
	ChatServerClientKey = "chat:client:"

	ServerConsumerKey = "chatserver:"

	OfflineMessageQueueKey = "offlineMessages"
)


const (
	PhoneCodeDuration = time.Minute * 3     // 手机验证码有效期
 	PhoneCodeResendDuration = time.Minute    // 手机验证码重复发送的间隔
	TokenExpireDuration = time.Second * 30
	UserCacheExpireDuration = time.Minute * 30
	DefaultCacheDuration
)
