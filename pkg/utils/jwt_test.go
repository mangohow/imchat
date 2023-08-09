package utils

import (
	"fmt"
	"testing"
)

/*
* @Author: mgh
* @Date: 2022/2/24 19:15
* @Desc:
 */

func TestToken(t *testing.T) {
	token, err := CreateToken(1, "1158446387")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(token)

	username, id, err := ParseToken(token)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(username)
	fmt.Println(id)
}


func TestMd5String(t *testing.T) {
	str := Md5String("123456")
	t.Log(str)
}