package localchannel

import (
	"sync"

	"github.com/FactomProject/factomd/modules/pubsub2"
)

type channel struct {
	channel chan interface{}

	reader *reader
	writer *writer

	close  sync.Once
	closed bool
	mtx    sync.RWMutex
}

var _ pubsub2.IChannel = (*channel)(nil)

func New(size int) *channel {
	lc := new(channel)
	lc.channel = make(chan interface{}, size)

	lc.reader = new(reader)
	lc.reader.c = lc.channel

	lc.writer = new(writer)
	lc.writer.c = lc.channel
	return lc
}

func (c *channel) GetWriter() pubsub2.IChannelWriter {
	return c.writer
}

func (c *channel) GetReader() pubsub2.IChannelReader {
	return c.reader
}

func (c *channel) Close() {
	c.close.Do(func() {
		c.mtx.Lock()
		c.closed = true
		close(c.channel)
		c.mtx.Unlock()
	})
}

func (c *channel) IsClosed() bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.closed
}
