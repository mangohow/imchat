package mq

import (
	"github.com/mangohow/imchat/cmd/chatserver/internal/conf"
	"github.com/mangohow/imchat/pkg/common/xmq"
	"github.com/streadway/amqp"
)



type MQProducer struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

type MQConsumer struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	DeliveryChan <-chan amqp.Delivery
}

var (
	ProducerInstance *MQProducer
	ConsumerInstance *MQConsumer
)

func InitMQ(consumerQueueName, consumerName string) error {
	conn, err := xmq.NewRabbitmqInstance(conf.MqConf)
	if err != nil {
		return err
	}

	producerChan, err := conn.Channel()

	ProducerInstance = &MQProducer{
		Conn:    conn,
		Channel: producerChan,
	}

	consumerChan, err := conn.Channel()
	if err != nil {
		return err
	}

	_, err = consumerChan.QueueDeclare(consumerQueueName,
		true,  // 是否持久化
		false, // 是否自动删除
		false, // 是否排他
		false, // nowait,
		nil,
	)
	if err != nil {
		return nil
	}

	ch, err := consumerChan.Consume(
		consumerQueueName,
		consumerName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	ConsumerInstance = &MQConsumer{
		Conn:         conn,
		Channel:      consumerChan,
		DeliveryChan: ch,
	}

	return nil
}

func (p *MQProducer) Publish(queName string, data []byte) error {
	return p.Channel.Publish("",
		queName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		})
}






