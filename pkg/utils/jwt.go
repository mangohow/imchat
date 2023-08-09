package utils

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

/*
* @Author: mgh
* @Date: 2022/2/24 18:59
* @Desc:
 */

type Claim struct {
	Username string
	UserId   int64
	jwt.StandardClaims
}

var jwtKey = []byte("imchat")

func CreateToken(id int64, username string) (string, error) {
	claims := &Claim{
		Username: username,
		UserId:   id,
		StandardClaims: jwt.StandardClaims{
			IssuedAt: time.Now().Unix(),
			Issuer:   "mgh",
			Subject:  "UserToken",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func ParseToken(token string) (int64, string, error) {
	if token == "" {
		return 0, "", errors.New("empty String")
	}
	data, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return 0, "", err
	}

	claim, ok := data.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", errors.New("parse Error")
	}

	return int64(claim["UserId"].(float64)), claim["Username"].(string), nil
}
