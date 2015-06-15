package strobe

import "sync"

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

//Strobe is an emitter that allows broadcasting and listening to messages via channels
type Strobe struct {
	listeners map[listener]bool
	sync.Mutex
}

//Listen creates a new receiver channel which acts as a subscription. In order to prevent leaks, always return a channel after use via `Forget`
func (s *Strobe) Listen() ClosableReceiver {
	l := listener{channel: make(chan string), closer: make(chan bool)}
	s.Lock()
	s.listeners[l] = true
	s.Unlock()
	go func() {
		<-l.closer
		s.Lock()
		delete(s.listeners, l)
		s.Unlock()
		close(l.channel)
		close(l.closer)
	}()
	return &l
}

//Pulse sends a message to all listening channels
func (s *Strobe) Pulse(message string) {
	s.Lock()
	for c := range s.listeners {
		go func(ch chan string, m string) {
			ch <- message
		}(c.channel, message)
	}
	s.Unlock()
}

//Count gives the number of listeners
func (s *Strobe) Count() int {
	return len(s.listeners)
}

//NewStrobe creates a new Strobe that can be used for PubSub
func NewStrobe() *Strobe {
	return &Strobe{listeners: make(map[listener]bool)}
}
