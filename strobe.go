package strobe

import "sync"

// ReceiveCloser provides a read only channel that needs to be closed after use.
type ReceiveCloser[T any] interface {
	Receiver() <-chan T
	Close()
}

// Get the channel on which messages are received
func (l *listener[T]) Receiver() <-chan T {
	return l.channel
}

// Close this receiver
func (l *listener[T]) Close() {
	close(l.closer)
}

// New creates a new Strobe.
func New[T any]() *Strobe[T] {
	return &Strobe[T]{
		listeners: make(map[listener[T]]struct{}),
		lock:      sync.Mutex{},
	}
}

// Listen creates a new ReceiveCloser on which Pulsed messages can be read.
// Close() must be called after usage.
func (s *Strobe[T]) Listen() ReceiveCloser[T] {
	l := listener[T]{channel: make(chan T), closer: make(chan struct{})}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.listeners[l] = struct{}{}
	go s.waitForClose(l)
	return &l
}

// Pulse sends a message to all listening channels created by Listen().
func (s *Strobe[T]) Pulse(message T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for l := range s.listeners {
		go s.send(l, message)
	}
}

// Count the number of active listeners on this Strobe.
func (s *Strobe[T]) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.listeners)
}

// Strobe is an emitter that allows broadcasting messages by channel fan-out.
type Strobe[T any] struct {
	listeners map[listener[T]]struct{}
	lock      sync.Mutex
}

type listener[T any] struct {
	channel chan T
	closer  chan struct{}
}

func (s *Strobe[T]) waitForClose(l listener[T]) {
	<-l.closer
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.listeners, l)
	close(l.channel)
}

func (s *Strobe[T]) send(l listener[T], message T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.listeners[l]; ok {
		l.channel <- message
	}
}

