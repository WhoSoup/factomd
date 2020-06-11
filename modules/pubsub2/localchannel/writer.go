package localchannel

import (
	"github.com/FactomProject/factomd/modules/pubsub2"
)

type writer struct {
	c *channel
}

var _ pubsub2.IChannelWriter = (*writer)(nil)

func (w *writer) Write(v interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = pubsub2.NewChannelIsClosedError()
		}
	}()

	w.c.channel <- v
	return nil
}
