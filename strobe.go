package strobe

import "sync"

// ClosableReceiver provides a read only channel that needs to be closed after use.
type ClosableReceiver interface {
	Receiver() <-chan string
	Close()
}

type listener struct {
	channel chan string
	closer  chan bool
}

func (l *listener) Receiver() <-chan string {
	return l.channel
}

func (l *listener) Close() {
	l.closer <- true
}

// Strobe is an emitter that allows broadcasting messages by channel fan-out.
type Strobe struct {
	listeners map[listener]bool
	lock      sync.Mutex
}

// Listen creates a new ClosableReceiver on which holds receiver channel on
// which messages can be received. Close() must be called after usage.
func (s *Strobe) Listen() ClosableReceiver {
	l := listener{channel: make(chan string), closer: make(chan bool)}
	s.lock.Lock()
	s.listeners[l] = true
	s.lock.Unlock()
	go func() {
		<-l.closer
		s.lock.Lock()
		delete(s.listeners, l)
		s.lock.Unlock()
		close(l.channel)
		close(l.closer)
	}()
	return &l
}

// Pulse sends a message to all listening channels
func (s *Strobe) Pulse(message string) {
	s.lock.Lock()
	for c := range s.listeners {
		go func(ch chan string, m string) {
			ch <- message
		}(c.channel, message)
	}
	s.lock.Unlock()
}

// Count the number of active listeners on this strobe
func (s *Strobe) Count() int {
	return len(s.listeners)
}

//NewStrobe creates a new Strobe that can be used for PubSub
func NewStrobe() *Strobe {
	return &Strobe{listeners: make(map[listener]bool)}
}
