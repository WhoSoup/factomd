package pubsub2

type IChannel interface {
	NewReader() IChannelReader
	NewWriter() IChannelWriter
	Close()
	IsClosed() bool
}

type IChannelReader interface {
	Reader() <-chan interface{}
	Read() (interface{}, bool)
}
type IChannelWriter interface {
	Write(interface{}) error
}

type ICallback func(interface{})
