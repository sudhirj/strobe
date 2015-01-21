package strobe

import "sync"

//Strobe is an emitter that allows broadcasting and listening to messages via channels
type Strobe struct {
	listeners map[chan string]bool
	views     map[<-chan string]chan string
	sync.RWMutex
}

//Listen creates a new receiver channel which acts as a subscription. In order to prevent leaks, always return a channel after use via `Forget`
func (s *Strobe) Listen() <-chan string {
	listener := make(chan string)
	s.Lock()
	s.listeners[listener] = true
	s.views[listener] = listener
	s.Unlock()
	return listener
}

//Pulse sends a message to all listening channels
func (s *Strobe) Pulse(message string) {
	s.Lock()
	for c := range s.listeners {
		c <- message
	}
	s.Unlock()
}

//Count gives the number of listeners
func (s *Strobe) Count() int {
	return len(s.listeners)
}

//Off removes a channel from the list of receivers
func (s *Strobe) Off(view <-chan string) {
	s.Lock()
	delete(s.listeners, s.views[view])
	delete(s.views, view)
	s.Unlock()
}

//NewStrobe creates a new Strobe that can be used for PubSub
func NewStrobe() *Strobe {
	return &Strobe{listeners: make(map[chan string]bool), views: make(map[<-chan string]chan string)}
}
