package mongodb

import (
	"github.com/mangohow/imchat/cmd/messageserver/internal/conf"
	"github.com/mangohow/imchat/pkg/common/xmongo"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	MongoClient *mongo.Client
	MongoDB *mongo.Database
)

func InitMongoDB() error {
	client, err := xmongo.NewMongoInstance(conf.MongoConf)
	if err != nil {
		return err
	}

	MongoClient = client

	MongoDB = MongoClient.Database(conf.MongoConf.Db)

	return nil
}
