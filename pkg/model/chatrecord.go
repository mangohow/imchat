package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 用户聊天记录

type ChatRecord struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Sender      int64              `json:"sender" bson:"sender"`
	Receiver    int64              `json:"receiver" bson:"receiver"`
	Message     []byte             `json:"message" bson:"message"`
	CreateTime  int64              `json:"createTime" bson:"createTime"`
	MessageType int32              `json:"messageType" bson:"messageType"`
	// 消息状态   0 未读  1 已读 2 发送方删除 3 接收方删除 4 双方删除
	Status      int32              `json:"status" bson:"status"`
}

const (
	RecordStatusUnread = iota
	RecordStatusRead
	RecordStatusReceiverRemoved
	RecordStatusSenderRemoved
	RecordStatusBothRemoved
)
