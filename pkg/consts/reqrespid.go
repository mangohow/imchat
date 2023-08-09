package consts


const (
	// 双向消息，客户端和服务端都可能发送这样的消息

	HelloRequest = iota + 1
	HelloReply


	SingleChatMessage


	SingleChatAck= iota + 10000



	// 单向消息，只有服务端会发送给客户端

	NewMessage = iota + 20000
)
