package handlers

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mangohow/imchat/cmd/chatserver/internal/chatserver"
	"github.com/mangohow/imchat/cmd/chatserver/internal/log"
	"github.com/mangohow/imchat/cmd/chatserver/internal/mongodb/dao"
	"github.com/mangohow/imchat/cmd/chatserver/internal/mq"
	"github.com/mangohow/imchat/cmd/chatserver/internal/rdsconn"
	"github.com/mangohow/imchat/pkg/consts/redisconsts"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/mangohow/imchat/proto/pb"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
)

type UserChatHandler struct {
	logger *logrus.Logger
	redis *redis.Client
	messageDao *dao.MessageDao
	retryHandler IRetryHandler
}

func NewUserChatHandler(ctx context.Context, worker int, retryHandler IRetryHandler) *UserChatHandler {
	h := &UserChatHandler{
		logger: log.Logger(),
		redis: rdsconn.RedisConn(),
		messageDao: dao.NewMessageDao("singleChat"),
		retryHandler: retryHandler,
	}

	for i := 0; i < worker; i++ {
		go h.forwardMessage(ctx)
	}

	return h
}

// ForwardMessage 转发用户消息
// 1. 如果Receiver在线
// 1.1 如果Receiver在同一服务器上，直接发送
// 1.2 否则，发送到消息队列中
// 2. 如果Receiver不在线，发送到messageServer的消息队列中
func (h *UserChatHandler) ForwardMessage(ctx *chatserver.Context, req *pb.SingleChat) *pb.ChatAck {
	// 检查参数合法性
	if !h.checkUserParam(ctx, req) {
		h.logger.Warning("user req parm invalid")
		return nil
	}

	// createTime 保存毫秒时间戳
	req.CreateTime = time.Now().UnixMicro()

	// 1.先将消息持久化消息到数据库中
	record := &model.ChatRecord{
		Sender:      req.Sender,
		Receiver:    req.Receiver,
		Message:     req.Message,
		CreateTime:  req.CreateTime,
		MessageType: int32(req.MsgType),
		Status:      model.RecordStatusUnread,
	}
	objId, err := h.messageDao.PersistUnreadMessage(record)
	if err != nil {
		h.logger.Errorf("persist message error:%v", err)
		return nil
	}
	req.MessageId = objId.Hex()

	// 回复
	ack := &pb.ChatAck{MessageSeq: req.MessageSeq, MessageId: req.MessageId}

	// 生成转发数据
	forwardData, err := proto.Marshal(req)
	if err != nil {
		h.logger.Errorf("marshal error:%v", err)
		return nil
	}

	forwardbuf := bytes.NewBuffer(nil)
	// 写入消息ID
	forwardbuf.Write(ctx.Message.RawData[:4])
	forwardbuf.Write(forwardData)

	h.logger.Debugf("[sender]:%d, [receiver]:%d, [message]:%s",
		req.Sender, req.Receiver, req.Message)

	// 2.检查是否在同一服务器，如果在直接发送
	err, ok := h.sendIfOnSameServer(forwardbuf.Bytes(), req)
	if err != nil {
		h.logger.Errorf("send error:%v", err)
		return nil
	}
	if ok {
		return ack
	}

	// 3.不在同一台服务器上，查询所在服务器，并发送到对应消息队列
	err, ok = h.sendIfOnAnotherServer(req, forwardbuf.Bytes())
	if err != nil {
		h.logger.Errorf("send async error:%v", err)
		return nil
	}
	if ok {
		return ack
	}

	// 4.用户不在线, 数据已经先被持久化到数据库中了
	// 待用户上线后主动拉取离线消息

	return ack
}

func (h *UserChatHandler) checkUserParam(ctx *chatserver.Context, req *pb.SingleChat) bool {
	// 不能发消息给自己
	if req.Sender == req.Receiver {
		return false
	}

	id, ok := ctx.GetInt64("id")
	if !ok {
		h.invalidOperation(ctx)
		return false
	}

	if req.Sender != id {
		h.invalidOperation(ctx)
		return false
	}

	// 检查是否是它的联系人，如果不是则不允许发送
	isMember, _ := h.redis.SIsMember(context.Background(), redisconsts.FriendsKey+strconv.Itoa(int(id)), req.Receiver).Result()
	if !isMember {
		h.invalidOperation(ctx)
		return false
	}

	// 检查消息序列是否合法, 消息序列由客户端生成, 用于标识一天内的唯一消息
	// 由32位秒时间戳和counter组成
	if req.MessageSeq >> 32 == 0 {
		h.invalidOperation(ctx)
		return false
	}

	return true
}

// 在同一服务器上
func (h *UserChatHandler) sendIfOnSameServer(data []byte, req *pb.SingleChat) (error, bool) {
	// 先查询是不是在同一台服务器上, 如果是直接转发
	target := chatserver.ClientManagerInstance.Get(req.Receiver)
	if target == nil {
		return nil, false
	}

	target.Write(data)
	h.retryHandler.Add(req.Receiver)

	return nil, true
}

// 在其它服务器上
func (h *UserChatHandler) sendIfOnAnotherServer(req *pb.SingleChat, data []byte) (error, bool) {
	result, err := h.redis.Get(context.Background(), redisconsts.ChatServerClientKey+strconv.Itoa(int(req.Receiver))).Result()
	if err == redis.Nil {
		// 用户不在线
		return nil, false
	}

	if err != nil {
		h.logger.Errorf("get client error:%v", err)
		return err, false
	}

	// 用户在线, 发送到消息队列
	queueName := redisconsts.ServerConsumerKey + result
	err = mq.ProducerInstance.Publish(queueName, data)
	if err != nil {
		h.logger.Errorf("publish message error:%v", err)
		return err, false
	}

	return nil, true
}

func (h *UserChatHandler) invalidOperation(ctx *chatserver.Context) {
	ctx.Abort()
	ctx.Close()
}

// forwardMessage 从消息队列中读取消息，并转发给用户
func (h *UserChatHandler) forwardMessage(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case delivery := <- mq.ConsumerInstance.DeliveryChan:
			if err := h.sendMessage(&delivery); err != nil {
				h.logger.Errorf("deliver message error:%v", err)
			}

			if err := delivery.Ack(false); err != nil {
				h.logger.Errorf("mq ack error:%v", err)
			}
		}
	}
}

// sendMessage 转发消息给客户端，同时消息会被写入mongo中
// 用户需要回应ack，以将mongo中的消息设置为已读
func (h *UserChatHandler) sendMessage(delivery *amqp.Delivery) error {
	data := delivery.Body
	msg := new(pb.SingleChat)
	err := proto.Unmarshal(data[4:], msg)
	if err != nil {
		return err
	}

	client := chatserver.ClientManagerInstance.Get(msg.Receiver)
	// 用户下线了
	if client == nil {
		return nil
	}

	client.Write(data)

	h.retryHandler.Add(msg.Receiver)

	return nil
}

// ConfirmMessage 来自客户端的确认消息
func (h *UserChatHandler) ConfirmMessage(ctx *chatserver.Context, ack *pb.ChatAck) {
	id, err := primitive.ObjectIDFromHex(ack.MessageId)
	if err != nil {
		h.logger.Errorf("get objid error:%v", err)
		return
	}
	err = h.messageDao.UpdateMessageRead(id)
	if err != nil {
		h.logger.Errorf("update message status error:%v", err)
		return
	}

	h.retryHandler.Remove(ctx.GetUid())
}