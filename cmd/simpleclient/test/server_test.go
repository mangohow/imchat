package test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/mangohow/imchat/cmd/simpleclient/client"
)


func BenchmarkServerSingle(b *testing.B) {
	rand.Seed(time.Now().UnixMicro())
	var users =  [][]string{
		{"mangohow", "123456"},
		{"iamtomaaa", "123456"},

	}
	clients := []*client.ChatClient{
		client.NewChatClient(":6387"),
		client.NewChatClient(":6388"),
	}

	data := randomBytes(100)

	for i, chatClient := range clients {
		chatClient.Test(users[i][0], users[i][1])
	}

	client := clients[0]
	id := client.GetFriends()[0].Userinfo.Id
	for i := 0; i < b.N; i++ {
		client.SendMessageTo(id, data)
	}


}



func randomBytes(n int) []byte {
	alpha := make([]byte, 0, 26)
	for a := byte('a'); a < 'z'; a++ {
		alpha = append(alpha, a)
	}

	res := make([]byte, n)
	for i := 0; i < n; i++ {
		m := rand.Intn(len(alpha))
		res[i] = alpha[m]
	}

	return res
}