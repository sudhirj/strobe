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
	s.Lock()
	defer s.Unlock()
	listener := make(chan string)
	s.listeners[listener] = true
	s.views[listener] = listener
	return listener
}

//Pulse sends a message to all listening channels
func (s *Strobe) Pulse(message string) {
	s.RLock()
	defer s.RUnlock()
	for c := range s.listeners {
		c <- message
	}
}

//Count gives the number of listeners
func (s *Strobe) Count() int {
	return len(s.listeners)
}

//Off removes a channel from the list of receivers
func (s *Strobe) Off(view <-chan string) {
	s.Lock()
	defer s.Unlock()
	delete(s.listeners, s.views[view])
	delete(s.views, view)
}

//NewStrobe creates a new Strobe that can be used for PubSub
func NewStrobe() *Strobe {
	return &Strobe{listeners: make(map[chan string]bool), views: make(map[<-chan string]chan string)}
}
