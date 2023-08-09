package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mangohow/imchat/cmd/simpleclient/client"
	"github.com/mangohow/imchat/pkg/utils"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:6387", "specify ws addr")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	var (
		username, password string
	)

	for {
		fmt.Println("请输入用户名：")
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Printf("error:%v", err)
			continue
		}
		username = string(line)
		break
	}

	for {
		fmt.Println("请输入密码：")
		line, _, err := reader.ReadLine()
		if err != nil {
			log.Printf("error:%v", err)
			continue
		}
		password = string(line)
		break
	}

	chatClient := client.NewChatClient(*addr)
	chatClient.Register()
	chatClient.Login(username, utils.Md5String(password))
}



