package dao

import (
	"context"

	"github.com/mangohow/imchat/cmd/chatserver/internal/mongodb"
	"github.com/mangohow/imchat/pkg/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageDao struct {
	mongo *mongo.Database
	collection *mongo.Collection
	collectionName string
}

func NewMessageDao(collectionName string) *MessageDao {
	return &MessageDao{
		collectionName: collectionName,
		mongo: mongodb.MongoDB,
		collection: mongodb.MongoDB.Collection(collectionName),
	}
}


// PersistUnreadMessage 持久化未读消息
func (d *MessageDao) PersistUnreadMessage(record *model.ChatRecord) (primitive.ObjectID, error) {
	res, err := d.collection.InsertOne(context.Background(), record)
	return res.InsertedID.(primitive.ObjectID), err
}

// UpdateMessageRead 设置消息为已读
func (d *MessageDao) UpdateMessageRead(id primitive.ObjectID) error {
	filter := bson.D{{"_id", id}}
	update := bson.D{{"$set", bson.D{{"status", model.RecordStatusRead}}}}
	_, err := d.collection.UpdateOne(context.Background(), filter, update)
	return err
}