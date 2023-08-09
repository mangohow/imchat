package model

import "time"

/*
* @Author: mgh
* @Date: 2022/2/26 18:20
* @Desc:
 */

type Userinfo struct {
	Id        int64    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Nickname  string    `json:"nickname" db:"nickname"`
	Avatar    string    `json:"avatar" db:"avatar"`
	Signature string    `json:"signature" db:"signature"`
	Gender    uint8     `json:"gender" db:"gender"`
	Birthday  time.Time `json:"birthday" db:"birthday"`
	Address   string    `json:"address" db:"address"`
	HomeAddr  string    `json:"home_addr" db:"home_addr"`
	PhoneNum  string    `json:"phone_num" db:"phone_num"`
	Email     string    `json:"email" db:"email"`
}

type UserDTO struct {
	Userinfo
	Username string `json:"username" db:"username"`
}
