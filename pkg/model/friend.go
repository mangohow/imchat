package model

import "time"

type Friend struct {
	Id         uint32    `json:"id" db:"id"`
	UserId     int64    `json:"user_id,omitempty" db:"user_id"` // 用户ID
	FriendId   int64    `json:"friend_id" db:"friend_id"`       // 用户朋友的userid
	Remark     string    `json:"remark" db:"remark"`             // 备注
	Auth       uint8     `json:"auth" db:"auth"`                 // 权限
	CreateTime time.Time `json:"create_time" db:"create_time"`
	Status     uint8     `json:"status" db:"status"`
}

type FriendDTO struct {
	Userinfo   *Userinfo  `json:"userinfo"`
	Remark     string    `json:"remark"`
	CreateTime time.Time `json:"createTime"`
	Auth       int       `json:"auth"`
}
