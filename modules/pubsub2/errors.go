package pubsub2

import "errors"

type ChannelIsClosed error

func NewChannelIsClosedError() ChannelIsClosed {
	return ChannelIsClosed(errors.New("channel is closed"))
}

func IsChannelIsClosedError(e error) bool {
	_, ok := e.(ChannelIsClosed)
	return ok
}
