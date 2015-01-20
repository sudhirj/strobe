package strobe

import "sync"

//Strobe is an emitter that allows broadcasting and listening to messages via channels
type Strobe struct {
	listeners map[chan string]bool
	views     map[<-chan string]chan string
	lock      sync.Locker
}

//Listen creates a new receiver channel which acts as a subscription. In order to prevent leaks, always return a channel after use via `Forget`
func (p *Strobe) Listen() <-chan string {
	listener := make(chan string)
	p.lock.Lock()
	p.listeners[listener] = true
	p.views[listener] = listener
	p.lock.Unlock()
	return listener
}

//Pulse sends a message to all listening channels
func (p *Strobe) Pulse(message string) {
	p.lock.Lock()
	for c := range p.listeners {
		c <- message
	}
	p.lock.Unlock()
}

//Off removes a channel from the list of receivers
func (p *Strobe) Off(view <-chan string) {
	p.lock.Lock()
	delete(p.listeners, p.views[view])
	delete(p.views, view)
	p.lock.Unlock()
}

//NewStrobe creates a new Strobe that can be used for PubSub
func NewStrobe() *Strobe {
	return &Strobe{listeners: make(map[chan string]bool), views: make(map[<-chan string]chan string), lock: &sync.Mutex{}}
}
