package model

import "time"

/*
* @Author: mgh
* @Date: 2022/2/24 12:42
* @Desc:
 */

type User struct {
	Id          int64    `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password" db:"password"`
	CreateTime  time.Time `json:"createTime" db:"create_time"`
	RemoveTime  time.Time `json:"removeTime" db:"remove_time"`
	BanedTime   time.Time `json:"banedTime" db:"baned_time"`
	ReleaseTime time.Time `json:"releaseTime" db:"release_time"`
	Status      int       `json:"status" db:"status"`
	PhoneNum    string    `json:"phoneNum" db:"phone_num"`
}


type UserRegister struct {
	Username string `json:"username" validate:"required,min=8,max=20"`
	Password string `json:"password" validate:"required,len=32"`
	Phone string `json:"phone" validate:"required,phone"`
	Code string `json:"code" validate:"required,len=6"`
}

type UserLogin struct {
	Username string `json:"username" validate:"required,min=8,max=20"`
	Password string `json:"password" validate:"required,len=32"`
}