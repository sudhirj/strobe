package strobe

import "sync"

// ClosableReceiver provides a read only channel that needs to be closed after use.
type ClosableReceiver interface {
	Receiver() <-chan string
	Close()
}

type listener struct {
	channel chan string
	closer  chan struct{}
}

func (l *listener) Receiver() <-chan string {
	return l.channel
}

func (l *listener) Close() {
	close(l.closer)
}

// Strobe is an emitter that allows broadcasting messages by channel fan-out.
type Strobe struct {
	listeners map[listener]struct{}
	lock      sync.Mutex
}

// Listen creates a new ClosableReceiver on which holds a channel on
// which messages can be received. Close() must be called after usage.
func (s *Strobe) Listen() ClosableReceiver {
	l := listener{channel: make(chan string), closer: make(chan struct{})}
	s.lock.Lock()
	s.listeners[l] = struct{}{}
	s.lock.Unlock()
	go func() {
		<-l.closer
		s.lock.Lock()
		delete(s.listeners, l)
		s.lock.Unlock()
		close(l.channel)
	}()
	return &l
}

// Pulse sends a message to all listening channels that have been checked out
// with Listen().
func (s *Strobe) Pulse(message string) {
	s.lock.Lock()
	for c := range s.listeners {
		go func(l listener, m string) {
			s.lock.Lock()
			if _, ok := s.listeners[l]; ok {
				l.channel <- m
			}
			s.lock.Unlock()
		}(c, message)
	}
	s.lock.Unlock()
}

// Count the number of active listeners on this strobe.
func (s *Strobe) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.listeners)
}

// NewStrobe creates a new Strobe.
func NewStrobe() *Strobe {
	return &Strobe{
		listeners: make(map[listener]struct{}),
		lock:      sync.Mutex{},
	}
}
