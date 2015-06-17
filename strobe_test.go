package strobe

import (
	"log"
	"sync"
	"testing"
	"time"
)

func TestPulse(t *testing.T) {
	strobe := NewStrobe()
	waiter := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		waiter.Add(1)
		listener := strobe.Listen()
		defer listener.Close()
		go func(t *testing.T, waiter *sync.WaitGroup, listener ClosableReceiver) {
			message := <-listener.Receiver()
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter, listener)
	}
	if strobe.Count() != 100 {
		t.Error("should be 100 listeners by now, got", strobe.Count(), "instead")
	}

	for i := 0; i < 2; i++ {
		go func(t *testing.T, waiter *sync.WaitGroup, stb *Strobe) {
			waiter.Add(1)
			message := <-stb.Listen().Receiver()
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter, strobe)
	}

	for i := 0; i < 2; i++ {
		go func(t *testing.T, waiter *sync.WaitGroup) {
			waiter.Add(1)
			message := <-strobe.Listen().Receiver()
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter)
	}

	forgottenListener := strobe.Listen()
	forgottenListener.Close()
	go func() {
		message := <-forgottenListener.Receiver()
		if message != "" {
			t.Error("should not have sent on this channel")
		}
	}()

	success := make(chan bool)
	go func() {
		waiter.Wait()
		success <- true
	}()

	go func() {
		<-time.After(25 * time.Millisecond)
		strobe.Pulse("PULSE")
	}()

	select {
	case <-success:
	case <-time.After(1 * time.Second):
		t.Error("No pulse received")
	}
}

func TestMessaging(t *testing.T) {
	strobe := NewStrobe()
	c1 := make(chan bool)
	go func() {
		message := <-strobe.Listen().Receiver()
		if message == "M1" {
			c1 <- true
		}
	}()

	strobe.Listen() // Creating a channel but not listening on it

	go func() {
		<-time.After(10 * time.Millisecond)
		strobe.Pulse("M1")
	}()
	go func() {
		<-time.After(1 * time.Second)
		t.Error("no message")
	}()
	<-c1
}

func Example() {
	strobe := NewStrobe()
	w := &sync.WaitGroup{}
	w.Add(3)
	go func() {
		message := <-strobe.Listen().Receiver()
		log.Println(message)
		w.Done()
	}()
	go func() {
		message := <-strobe.Listen().Receiver()
		log.Println(message)
		w.Done()
	}()
	go func() {
		message := <-strobe.Listen().Receiver()
		log.Println(message)
		w.Done()
	}()

	strobe.Pulse("PING")
	w.Wait()

	// Output:
	// PING
	// PING
	// PING
}
