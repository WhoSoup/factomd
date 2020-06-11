package pubsub2

import (
	"errors"
	"sync"
)

var globalRegistry *Registry

func init() {
	Reset()
}

func Reset() {
	globalRegistry = new(Registry)
}

type Registry struct {
	mtx   sync.RWMutex
	paths map[string]IChannel
}

func (r *Registry) Register(path string, c IChannel) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.paths[path]; ok {
		return errors.New("channel already exists")
	}
	r.paths[path] = c
	return nil
}

func (r *Registry) Unregister(path string) {
	r.mtx.Lock()
	delete(r.paths, path)
	r.mtx.Unlock()
}

func (r *Registry) Get(path string) (IChannel, bool) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	v, ok := r.paths[path]
	return v, ok
}
