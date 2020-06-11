package localchannel

import (
	"github.com/FactomProject/factomd/modules/pubsub2"
)

type reader struct {
	c chan interface{}
}

var _ pubsub2.IChannelReader = (*reader)(nil)

func (r *reader) Channel() <-chan interface{} {
	return r.c
}
func (r *reader) Read() (interface{}, bool) {
	v, ok := <-r.c
	return v, ok
}
