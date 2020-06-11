package localchannel

import (
	"github.com/FactomProject/factomd/modules/pubsub2"
)

type reader struct {
	c *channel
}

var _ pubsub2.IChannelReader = (*reader)(nil)

func (r *reader) Reader() <-chan interface{} {
	return r.c.channel
}
func (r *reader) Read() (interface{}, bool) {
	v, ok := <-r.c.channel
	return v, ok
}
