package xmongo

import (
	"context"
	"time"

	"github.com/mangohow/imchat/pkg/common/xconfig"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoInstance(conf *xconfig.MongoConfig) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(conf.Url).
		SetMaxPoolSize(uint64(conf.MaxPoolSize)).
		SetMinPoolSize(uint64(conf.MinPoolSize))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	conn, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	return conn, err
}