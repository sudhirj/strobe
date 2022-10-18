package strobe

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestPulse(t *testing.T) {
	strobe := New[string]()
	waiter := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		waiter.Add(1)
		listener := strobe.Listen()
		defer listener.Close()
		go func(t *testing.T, waiter *sync.WaitGroup, listener ReceiveCloser[string]) {
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
		waiter.Add(1)
		go func(t *testing.T, waiter *sync.WaitGroup, stb *Strobe[string]) {
			message := <-stb.Listen().Receiver()
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter, strobe)
	}

	for i := 0; i < 2; i++ {
		waiter.Add(1)
		go func(t *testing.T, waiter *sync.WaitGroup) {
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

func TestRaceConditions(t *testing.T) {
	strobe := New[string]()
	go func() {
		for index := 0; index < 1000; index++ {
			go strobe.Count()
		}
	}()
	go func() {
		for index := 0; index < 1000; index++ {
			l := strobe.Listen()
			if math.Remainder(float64(index), 2.0) == 0 {
				go l.Close()
			} else {
				defer l.Close()
			}
		}
	}()
	go func() {
		for index := 0; index < 1000; index++ {
			go strobe.Pulse(strconv.Itoa(index))
		}
	}()
}

func TestMessaging(t *testing.T) {
	strobe := New[string]()

	strobe.Listen() // Creating a channel but not listening on it
	readySignal := make(chan struct{})
	go func() {
		<-readySignal
		strobe.Pulse("M1")
	}()

	select {
	case message := <-strobe.Listen().Receiver():
		if message != "M1" {
			t.Error("wrong message received")
		}
	case <-time.After(10 * time.Second):
		t.Error("no message received")
	case readySignal <- struct{}{}:
		// Signals the sender that the select case is ready
	}
}

func Example() {
	s := New[string]()
	w := &sync.WaitGroup{}
	w.Add(3)

	go func(listener ReceiveCloser[string]) {
		message := <-listener.Receiver()
		fmt.Println(message)
		listener.Close()
		w.Done()
	}(s.Listen())
	go func(listener ReceiveCloser[string]) {
		message := <-listener.Receiver()
		fmt.Println(message)
		listener.Close()
		w.Done()
	}(s.Listen())
	go func(listener ReceiveCloser[string]) {
		message := <-listener.Receiver()
		fmt.Println(message)
		listener.Close()
		w.Done()
	}(s.Listen())

	s.Pulse("PING")
	w.Wait()

	// Output:
	// PING
	// PING
	// PING
}
