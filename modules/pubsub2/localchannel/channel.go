package localchannel

import (
	"sync"

	"github.com/FactomProject/factomd/modules/pubsub2"
)

type channel struct {
	channel chan interface{}
	close   sync.Once
	closed  bool
	mtx     sync.RWMutex
}

var _ pubsub2.IChannel = (*channel)(nil)

func New(size int) *channel {
	lc := new(channel)
	lc.channel = make(chan interface{}, size)
	return lc
}

func (c *channel) NewWriter() pubsub2.IChannelWriter {
	w := new(writer)
	w.c = c
	return w
}

func (c *channel) NewReader() pubsub2.IChannelReader {
	r := new(reader)
	r.c = c
	return r
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
