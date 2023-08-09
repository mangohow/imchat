package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/gorilla/websocket"
	"github.com/mangohow/imchat/pkg/consts"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/mangohow/imchat/pkg/utils"
	"github.com/mangohow/imchat/proto/pb"
	"google.golang.org/protobuf/proto"
)

type ChatClient struct {
	httpClient *http.Client
	wsConn *websocket.Conn
	heartBeat time.Duration

	ctx context.Context
	cancelFunc context.CancelFunc

	stdInReader *bufio.Reader

	user model.User
	token string

	friends []model.FriendDTO
	friendsMap map[int64]model.FriendDTO

	messageHandler *MessageHandler

	wsAddr string

	messageCounter uint32

	writeMux sync.Mutex
}

func NewChatClient(addr string) *ChatClient {
	return &ChatClient{
		httpClient: &http.Client{},
		heartBeat: time.Second * 5,
		stdInReader: bufio.NewReader(os.Stdin),
		friendsMap: make(map[int64]model.FriendDTO),
		messageHandler: NewMessageHandler(),
		wsAddr: addr,
	}
}

func (c *ChatClient) GetFriends() []model.FriendDTO {
	return c.friends
}

func (c *ChatClient) Register() {
	c.messageHandler.Register(consts.SingleChatMessage, c.HandleSingleChatMessage)
	c.messageHandler.Register(consts.SingleChatAck, c.HandleSingleChatAck)
	c.messageHandler.Register(consts.NewMessage, c.HandleNewMessage)
}

func (c *ChatClient) Test(username, password string) {
	password = utils.Md5String(password)
	token, err := c.loginHttp(username, password)
	if err != nil {
		panic(err)
	}

	c.token = token
	id, username, err := utils.ParseToken(token)
	if err != nil {
		panic(err)
	}
	c.user.Id = id
	c.user.Username = username

	err = c.loginWebSocket(token)
	if err != nil {
		panic(err)
	}

	c.friends = c.getFriends()
	for i := range c.friends {
		c.friendsMap[c.friends[i].Userinfo.Id] = c.friends[i]
	}

	c.pullOfflineMessage()

	go c.startReader()
}

func (c *ChatClient) Login(username, password string) {
	token, err := c.loginHttp(username, password)
	if err != nil {
		panic(err)
	}

	c.token = token
	id, username, err := utils.ParseToken(token)
	if err != nil {
		panic(err)
	}
	c.user.Id = id
	c.user.Username = username

	err = c.loginWebSocket(token)
	if err != nil {
		panic(err)
	}

	c.friends = c.getFriends()
	for i := range c.friends {
		c.friendsMap[c.friends[i].Userinfo.Id] = c.friends[i]
	}

	c.pullOfflineMessage()

	log.Printf("[self id] %d", c.user.Id)
	c.HandleMessage()
}

func (c *ChatClient) getFriends() []model.FriendDTO {
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/auth/friends", nil)
	if err != nil {
		log.Printf("new request error:%v", err)
		return nil
	}
	request.Header.Set("authorization", c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		log.Printf("do http request error:%v", err)
		return nil
	}
	data, _ := io.ReadAll(response.Body)
	response.Body.Close()
	var f Response[[]model.FriendDTO]
	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Printf("marshal json error:%v", err)
		return nil
	}

	return f.Data
}

func (c *ChatClient) getOnlineFriends() []int64 {
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/auth/onlineFriends", nil)
	if err != nil {
		log.Printf("new request error:%v", err)
		return nil
	}
	request.Header.Set("authorization", c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		log.Printf("do http request error:%v", err)
		return nil
	}
	data, _ := io.ReadAll(response.Body)
	response.Body.Close()
	var f Response[[]int64]
	err = json.Unmarshal(data, &f)
	if err != nil {
		log.Printf("marshal json error:%v", err)
		return nil
	}

	return f.Data
}


func (c *ChatClient) loginHttp(username, password string) (string, error) {
	user := &model.UserLogin{
		Username: username,
		Password: password,
	}

	data, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("marshal user error:%v", err)
	}

	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/login", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("new request error:%v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("get response error:%v", err)
	}


	n := response.ContentLength
	buf := make([]byte, n)
	_, err = response.Body.Read(buf)
	defer response.Body.Close()
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read body error:%v", err)
	}

	resp := new(Response[LoginData])
	err = json.Unmarshal(buf, resp)
	if err != nil {
		return "", fmt.Errorf("unmarshal json error:%v", err)
	}

	return resp.Data.Token, nil
}


func (c *ChatClient) loginWebSocket(token string) error {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(fmt.Sprintf("ws://%s/ws/chat", c.wsAddr), map[string][]string{
		"authorization": {token},
	})
	if err != nil {
		return fmt.Errorf("dial websocket error:%v", err)
	}
	c.wsConn = conn

	c.ctx, c.cancelFunc = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second*10))
				if err != nil {
					panic(err)
				}
			case <- ctx.Done():
				return
			}
		}
	}(c.ctx)

	return nil
}

func (c *ChatClient) startReader() {
	for {
		_, p, err := c.wsConn.ReadMessage()
		if err != nil {
			log.Printf("read message error:%v", err)
		}

		id := binary.LittleEndian.Uint32(p[:4])
		if len(p) == 4 {
			c.messageHandler.Handle(id,nil)
		} else {
			c.messageHandler.Handle(id, p[4:])
		}

	}
}

func (c *ChatClient) HandleMessage() {
	defer c.cancelFunc()

	go c.startReader()

	// c.sendHello()
	for {
		fmt.Printf("请输入需要的操作：\n1.显示联系人\n2.发送消息\n3.查看历史消息\n4.清屏\n5.退出\n")
		str, err := c.stdInReader.ReadBytes('\n')
		if err != nil {
			panic(err)
		}
		str = bytes.Trim(str, "\r\n")
		idx, err := strconv.Atoi(string(str))
		if err != nil {
			log.Printf("输入格式有误：%v", err)
			continue
		}
		switch idx {
		case 1:
			c.handleShowFriends()
		case 2:
			c.handleSendMessage()
		case 3:
			c.showHistoryMessage()
		case 4:
			exec.Command("clear")
			exec.Command("cmd", "/c", "clear")
		case 5:
			os.Exit(0)
		}

	}
}

func (c *ChatClient) handleSendMessage() {
	fmt.Println("输入接收者索引号：")
	str, _ := c.stdInReader.ReadString('\n')
	idx, _ := strconv.Atoi(str)
	fmt.Println("输入消息：")
	message, _ := c.stdInReader.ReadBytes('\n')
	message = bytes.Trim(message, "\r\n")
	receiver := c.friends[idx].Userinfo.Id

	seq := time.Now().Unix()
	seq <<= 32
	seq |= int64(c.messageCounter)
	c.messageCounter++

	msg := &pb.SingleChat{
		MessageSeq: seq,
		Sender:   c.user.Id,
		Receiver: receiver,
		MsgType:  pb.MsgType_Text,
		Message:  message,
	}

	err := c.WriteProtoMessage(consts.SingleChatMessage, msg)
	if err != nil {
		log.Printf("write proto message error:%v", err)
	}
}

func (c *ChatClient) SendMessageTo(id int64, message []byte) {
	seq := time.Now().Unix()
	seq <<= 32
	seq |= int64(c.messageCounter)
	c.messageCounter++

	msg := &pb.SingleChat{
		MessageSeq: seq,
		Sender:   c.user.Id,
		Receiver: id,
		MsgType:  pb.MsgType_Text,
		Message:  message,
	}

	c.WriteProtoMessage(consts.SingleChatMessage, msg)
}

func (c *ChatClient) WriteProtoMessage(id uint32, message proto.Message) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, id)
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	data = append(buf, data...)

	c.writeMux.Lock()
	defer c.writeMux.Unlock()

	return c.wsConn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *ChatClient) handleShowFriends() {
	if len(c.friends) == 0 {
		return
	}

	onlineFriends := c.getOnlineFriends()
	m := map[int64]struct{}{}
	for i := range onlineFriends {
		m[onlineFriends[i]] = struct{}{}
	}

	for i, friend := range c.friends {
		fmt.Printf("[%d] id:%d\tusername:%s\tremark:%s",
			i, friend.Userinfo.Id, friend.Userinfo.Username, friend.Remark)
		if _, ok := m[friend.Userinfo.Id]; ok {
			fmt.Println("\t[online]")
		} else {
			fmt.Println()
		}

	}
}

func (c *ChatClient) sendHello() {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, consts.HelloRequest)
	hello := &pb.Hello{Message: "hello, chat server"}

	buffer := bytes.NewBuffer(buf)
	helloData, err := proto.Marshal(hello)
	if err != nil {
		panic(err)
	}
	buffer.Write(helloData)

	c.writeMux.Lock()
	defer c.writeMux.Unlock()
	err = c.wsConn.WriteMessage(websocket.BinaryMessage, buffer.Bytes())
	if err != nil {
		log.Printf("send hello error")
	}

}

// 拉取离线消息
func (c *ChatClient) pullOfflineMessage() {
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8081/api/message/offline", nil)
	if err != nil {
		log.Printf("new request error:%v", err)
		return
	}
	request.Header.Set("authorization", c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		log.Printf("do http error:%v", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("get offline message errror:%v", response.Status)
		return
	}
	var messages Response[map[int64][]model.ChatRecord]
	if response.ContentLength <= 0 {
		return
	}

	buf := make([]byte, response.ContentLength)
	response.Body.Read(buf)
	response.Body.Close()
	err = json.Unmarshal(buf, &messages)
	if err != nil {
		log.Printf("marshal json error:%v", err)
		return
	}

	if messages.Code != 0 {
		log.Printf("get offline message failed:%s", messages.Message)
		return
	}

	if len(messages.Data) == 0 {
		return
	}

	msgIds := make([]string, 0)

	for _, msgs := range messages.Data {
		pie.SortUsing(msgs, func(a, b model.ChatRecord) bool {
			return a.CreateTime < b.CreateTime
		})
		for _, msg := range msgs {
			msgIds = append(msgIds, msg.Id.Hex())
		}

	}

	c.updateOfflineMessageStatus(msgIds)

	c.PrintMessagesMap(messages.Data)
}

// 更新离线消息为已读
func (c *ChatClient) updateOfflineMessageStatus(ids []string) {
	reqData, err := json.Marshal(ids)
	if err != nil {
		log.Printf("marshal json error:%v", err)
		return
	}
	reader := bytes.NewReader(reqData)
	request, err := http.NewRequest(http.MethodPut, "http://127.0.0.1:8081/api/message/status", reader)
	if err != nil {
		log.Printf("new request error")
		return
	}
	request.Header.Set("authorization", c.token)
	response, err := c.httpClient.Do(request)
	if err != nil {
		log.Printf("do http error:%v", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("get offline message errror:%v", response.Status)
		return
	}
}

func (c *ChatClient) PrintMessagesMap(msgs map[int64][]model.ChatRecord) {
	for sender, msgs := range msgs {
		friend, ok := c.friendsMap[sender]
		if !ok {
			log.Printf("can not find friend, sender:%d", sender)
			return
		}
		fmt.Printf("------------SenderId:%d---------------\n", sender)
		for _, msg := range msgs {
			t := time.UnixMicro(msg.CreateTime).Format(time.DateTime)
			fmt.Printf("[Sender:%s %s] %s\n", friend.Remark, t, msg.Message)
		}
		fmt.Println("------------end---------------")
	}

}

func (c *ChatClient) PrintMessagesSlice(msgs []model.ChatRecord) {
	for _, msg := range msgs {
		var name string
		// 自己发的消息
		if msg.Sender == c.user.Id {
			name = "self"
		} else {
			friend, ok := c.friendsMap[msg.Sender]
			if !ok {
				log.Printf("can not find friend, sender:%d", msg.Sender)
				return
			}
			name = friend.Remark
		}


		t := time.UnixMicro(msg.CreateTime).Format(time.DateTime)
		fmt.Printf("[Sender:%s %s] %s\n", name, t, msg.Message)
	}
}

func (c *ChatClient) PrintMessageProto(msg *pb.SingleChat) {
	friend, ok := c.friendsMap[msg.Sender]
	if !ok {
		log.Printf("can not find friend, sender:%d", msg.Sender)
		return
	}
	t := time.UnixMicro(msg.CreateTime).Format(time.DateTime)
	fmt.Printf("[Sender:%s %s] %s\n", friend.Remark, t, msg.Message)
}

func (c *ChatClient) showHistoryMessage() {
	fmt.Println("请输入要查看的朋友的索引：")
	str, err := c.stdInReader.ReadString('\n')
	if err != nil {
		log.Printf("read stdin error:%v", err)
		return
	}
	str = strings.Trim(str, "\r\n")
	idx, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("请输入索引（数字）")
		return
	}

	friend := c.friends[idx].Userinfo.Id
	pageSize := 5
	createTime := int64(-1)

	for {
		url := fmt.Sprintf("http://127.0.0.1:8081/api/message/history?friendId=%d&pageSize=%d&createTime=%d",
			friend, pageSize, createTime)
		request, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Printf("new request error:%v", err)
			return
		}
		request.Header.Set("authorization", c.token)
		response, err := c.httpClient.Do(request)
		if err != nil {
			log.Printf("do http request error:%v", err)
			return
		}
		buf := make([]byte, response.ContentLength)
		response.Body.Read(buf)
		response.Body.Close()
		var messages Response[[]model.ChatRecord]
		err = json.Unmarshal(buf, &messages)
		if err != nil {
			log.Printf("unmarshal json error:%v", err)
			return
		}

		if len(messages.Data) == 0 {
			fmt.Println("没有更多消息了")
			return
		}

		createTime = messages.Data[len(messages.Data)-1].CreateTime
		c.PrintMessagesSlice(pie.Reverse(messages.Data))

		fmt.Println("继续查看：n, 结束：q")
		str, err := c.stdInReader.ReadString('\n')
		if err != nil {
			log.Printf("read stdin error:%v", err)
			continue
		}
		str = strings.Trim(str, "\r\n")
		switch str {
		case "n":
			exec.Command("clear")
			exec.Command("cmd", "/c", "clear")
		case "q":
			return
		}
	}
}