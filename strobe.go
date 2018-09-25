package strobe

import (
	"sync"
	"time"
	"log"
)

const sendTimeout = 1 * time.Second

// ClosableReceiver provides a read only channel that needs to be closed after use.
type ClosableReceiver interface {
	Receiver() <-chan string
	Close()
}

// Get the channel on which messages are received
func (l *listener) Receiver() <-chan string {
	return l.channel
}

// Close this receiver
func (l *listener) Close() {
	close(l.closer)
}

// NewStrobe creates a new Strobe.
func NewStrobe() *Strobe {
	return &Strobe{
		listeners: make(map[listener]struct{}),
		lock:      sync.Mutex{},
	}
}

// Listen creates a new ClosableReceiver on which holds a channel on
// which messages can be received. Close() must be called after usage.
func (s *Strobe) Listen() ClosableReceiver {
	l := listener{channel: make(chan string), closer: make(chan struct{})}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.listeners[l] = struct{}{}
	go s.waitForClose(l)
	return &l
}

// Pulse sends a message to all listening channels that have been checked out
// with Listen().
func (s *Strobe) Pulse(message string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for l := range s.listeners {
		go s.send(l, message)
	}
}

// Count the number of active listeners on this Strobe.
func (s *Strobe) Count() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.listeners)
}

// Strobe is an emitter that allows broadcasting messages by channel fan-out.
type Strobe struct {
	listeners map[listener]struct{}
	lock      sync.Mutex
}

type listener struct {
	channel chan string
	closer  chan struct{}
}

func (s *Strobe) waitForClose(l listener) {
	<-l.closer
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.listeners, l)
	close(l.channel)
}

func (s *Strobe) send(l listener, message string) {
    s.lock.Lock()
    defer s.lock.Unlock()
    t := time.NewTimer(sendTimeout)

    if _, ok := s.listeners[l]; ok {
        select {
        case l.channel <- message:
        default:
            t.Reset(sendTimeout)
            select {
                    case l.channel <- message:
                    case <-t.C:
                        log.Printf("send timed out on %v\n", l)
            }
            if !t.Stop() {
                <-t.C
            }
        }
    }
}
