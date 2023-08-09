package xmq

import (
	"fmt"

	"github.com/mangohow/imchat/pkg/common/xconfig"
	"github.com/streadway/amqp"
)

func NewRabbitmqInstance(conf *xconfig.RabbitMqConfig) (*amqp.Connection, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		conf.Username, conf.Password, conf.Host, conf.Port)
	return amqp.Dial(url)
}

