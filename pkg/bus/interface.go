package bus

type Streamer interface {
	PublishOutbound(message OutboundMessage) // 流失输出
}
