package utils

import (
	"crypto/md5"
	"fmt"
)

func Md5String(str string) string {
	return fmt.Sprintf("%x", md5.Sum(String2Bytes(str)))
}
