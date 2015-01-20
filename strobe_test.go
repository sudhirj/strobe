package strobe

import (
	"sync"
	"testing"
	"time"
)

func TestStrobe(t *testing.T) {
	strobe := NewStrobe()
	waiter := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		waiter.Add(1)
		listener := strobe.Listen()
		go func(t *testing.T, waiter *sync.WaitGroup, listener <-chan string) {
			t.Log("adding listener")
			message := <-listener
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter, listener)
	}
	forgottenListener := strobe.Listen()
	strobe.Off(forgottenListener)
	go func() {
		<-forgottenListener
		t.Error("should not have sent on this channel")
	}()

	success := make(chan bool)
	go func() {
		waiter.Wait()
		success <- true
	}()

	strobe.Pulse("PULSE")

	select {
	case <-success:
	case <-time.After(2 * time.Second):
		t.Error("No pulse received")
	}
}
