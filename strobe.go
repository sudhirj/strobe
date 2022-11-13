package strobe

import (
	"context"
	"sync"
)

// Strobe is an emitter that allows broadcasting messages by channel fan-out.
type Strobe[T any] struct {
	listeners map[chan T]struct{}
	lock      sync.Mutex
}

// New creates a new Strobe.
func New[T any]() *Strobe[T] {
	return &Strobe[T]{
		listeners: make(map[chan T]struct{}),
		lock:      sync.Mutex{},
	}
}

// Listener returns a channel which receives pulsed messages from the strobe.
// The channel will be closed when ctx is cancelled.
func (s *Strobe[T]) Listener(ctx context.Context) <-chan T {
	l := make(chan T)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.listeners[l] = struct{}{}
	go s.waitForClose(ctx, l)
	return l
}

func (s *Strobe[T]) waitForClose(ctx context.Context, ch chan T) {
	<-ctx.Done()
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.listeners, ch)
	close(ch)
}

// Pulse sends a message to all listening channels created by Listen().
func (s *Strobe[T]) Pulse(message T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for l := range s.listeners {
		go s.send(l, message)
	}
}

func (s *Strobe[T]) send(l chan<- T, message T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	l <- message
}

// Count the number of active listeners on this Strobe.
func (s *Strobe[T]) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.listeners)
}
