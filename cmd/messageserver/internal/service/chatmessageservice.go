package service

import (
	"context"

	"github.com/mangohow/imchat/cmd/messageserver/internal/log"
	"github.com/mangohow/imchat/cmd/messageserver/internal/mongodb"
	"github.com/mangohow/imchat/pkg/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatMessageService struct {
	db *mongo.Collection
	logger *logrus.Logger
}


func NewChatMessageService() *ChatMessageService {
	return &ChatMessageService{
		db: mongodb.MongoDB.Collection("singleChat"),
		logger: log.Logger(),
	}
}


func (s *ChatMessageService) UpdateMany(uid int64, msgIds []string) (err error) {
	writeModel :=make([]mongo.WriteModel, len(msgIds))
	update := bson.D{{"$set", bson.D{{"status", model.RecordStatusRead}}}}
	for i := range msgIds {
		objId, _ := primitive.ObjectIDFromHex(msgIds[i])
		filter := bson.D{{"_id", objId}, {"receiver", uid}}
		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
		writeModel[i] = model
	}

	_, err = s.db.BulkWrite(context.Background(), writeModel)
	if err != nil {
		return err
	}

	return nil
}

func (s *ChatMessageService) UpdateOne(uid int64, msgId string) (err error) {
	objId, _ := primitive.ObjectIDFromHex(msgId)
	filter := bson.D{{"_id", objId}, {"receiver", uid}}
	update := bson.D{{"$set", bson.D{{"status", model.RecordStatusRead}}}}
	_, err = s.db.UpdateOne(context.Background(), filter, update)
	return
}

func (s *ChatMessageService) GetMessage(id int64, friendId int64, pageSize int, createTime int64) (records []model.ChatRecord, err error) {
	// 查找最新聊天记录
	filter := bson.M{
		"$or": bson.A{
			bson.M{"receiver": id, "sender": friendId},
			bson.M{"receiver": friendId, "sender": id},
		},
	}
	/*
	db.singleChat.find({
	  createTime: { $lt: 1691503996133385 },
	  $or: [
	    { $and: [{ sender: 110846059347969 }, { receiver: 110846065770498 }] },
	    { $and: [{ sender: 110846065770498 }, { receiver: 110846059347969 }] }
	  ]
	})
	*/
	// 根据位置查询最新聊天记录
	if createTime != -1 {
		filter["createTime"] = bson.M{"$lt": createTime}
	}
	// 按照createTime降序排序，获取最新数据 1 为升序 2为降序
	sort := bson.D{{"createTime", -1}}
	opts := options.Find().SetLimit(int64(pageSize)).SetSort(sort)
	res, err := s.db.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	err = res.All(context.Background(), &records)

	return
}

func (s *ChatMessageService) GetOfflineMessage(id int64) (recs map[int64][]*model.ChatRecord, err error) {
	filter := bson.D{{"receiver", id}, {"status", model.RecordStatusUnread}}
	res, err := s.db.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var records []model.ChatRecord
	err = res.All(context.Background(), &records)
	if err != nil || len(records) == 0 {
		return nil, err
	}

	recs = make(map[int64][]*model.ChatRecord)
	for i := range records {
		recs[records[i].Sender] = append(recs[records[i].Sender], &records[i])
	}

	return
}
