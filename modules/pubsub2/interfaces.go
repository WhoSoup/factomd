package pubsub2

type IChannel interface {
	GetReader() IChannelReader
	GetWriter() IChannelWriter
	Close()
	IsClosed() bool
}

type IChannelReader interface {
	Channel() <-chan interface{}
	Read() (interface{}, bool)
}
type IChannelWriter interface {
	Write(interface{}) error
}

type ICallback func(interface{})
