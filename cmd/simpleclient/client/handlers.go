package client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/mangohow/imchat/pkg/consts"
	"github.com/mangohow/imchat/proto/pb"
	"google.golang.org/protobuf/proto"
)

func (c *ChatClient) HandleSingleChatMessage(data []byte) {
	message := new(pb.SingleChat)
	err := proto.Unmarshal(data, message)
	if err != nil {
		log.Printf("proto marshal error:%v", err)
		return
	}

	// 确认消息
	ack := &pb.ChatAck{
		MessageSeq: message.MessageSeq,
		MessageId:  message.MessageId,
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, consts.SingleChatAck)
	buffer := bytes.Buffer{}
	buffer.Write(buf)

	ackData, err := proto.Marshal(ack)
	if err != nil {
		log.Printf("marshal error:%v", err)
	}
	buffer.Write(ackData)
	c.wsConn.WriteMessage(websocket.BinaryMessage, buffer.Bytes())

	c.PrintMessageProto(message)
}

func (c *ChatClient) HandleSingleChatAck(data []byte) {
	ack := new(pb.ChatAck)
	err := proto.Unmarshal(data, ack)
	if err != nil {
		log.Printf("proto marshal error:%v", err)
		return
	}

	fmt.Printf("[server received:%d]\n", ack.MessageSeq)
}

func (c *ChatClient) HandleNewMessage(data []byte) {
	log.Printf("new message, need to pull offline")
	c.pullOfflineMessage()
}