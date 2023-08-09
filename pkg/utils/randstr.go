package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)


func GetRandomCode(n int) string {
	builder := strings.Builder{}
	builder.Grow(n)
	for i := 0; i < n; i++ {
		num := rand.Intn(10)
		builder.WriteString(strconv.Itoa(num))
	}

	return builder.String()
}

var alphs = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k',
	'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K',
	'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}

func GetRandomAlphString(n int) string {
	builder := strings.Builder{}
	builder.Grow(n)
	l := len(alphs)
	for i := 0; i < n; i++ {
		idx := rand.Intn(l)
		builder.WriteByte(alphs[idx])
	}

	return builder.String()
}



func init() {
	rand.Seed(time.Now().UnixNano())
}