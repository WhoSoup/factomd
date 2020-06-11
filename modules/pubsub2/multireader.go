package pubsub2

import "sync"

// order not guaranteed
type MultiReader struct {
	channel IChannel

	listeners []chan interface{}
	callbacks []ICallback
	mtx       sync.RWMutex

	close chan interface{}
}

func NewMultiReader(channel IChannel) *MultiReader {
	mr := new(MultiReader)
	mr.channel = channel
	mr.close = make(chan interface{})
	go mr.listen()
	return mr
}

func (mr *MultiReader) NewListener() <-chan interface{} {
	mr.mtx.Lock()
	c := make(chan interface{})
	mr.listeners = append(mr.listeners, c)
	mr.mtx.Unlock()
	return c
}

func (mr *MultiReader) RemoveListener(c <-chan interface{}) {
	mr.mtx.Lock()
	for i := range mr.listeners {
		if mr.listeners[i] == c {
			mr.listeners = append(mr.listeners[:i], mr.listeners[i+1:]...)
			return
		}
	}
	mr.mtx.Unlock()
}

func (mr *MultiReader) NewCallback(c ICallback) {
	mr.mtx.Lock()
	mr.callbacks = append(mr.callbacks, c)
	mr.mtx.Unlock()
}

func (mr *MultiReader) Close() {
	mr.mtx.Lock()
	close(mr.close)
	for i := range mr.listeners {
		close(mr.listeners[i])
	}
	mr.listeners = nil
	mr.callbacks = nil
	mr.mtx.Unlock()
}

func (mr *MultiReader) listen() {
	reader := mr.channel.GetReader().Channel()
	for {
		select {
		case <-mr.close:
			return
		case v := <-reader:
			mr.mtx.RLock()
			for i := range mr.listeners {
				mr.listeners[i] <- v
			}
			for i := range mr.callbacks {
				mr.callbacks[i](v)
			}
			mr.mtx.RUnlock()
		}
	}
}
