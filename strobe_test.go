package strobe

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPulse(t *testing.T) {
	strobe := New[string]()
	waiter := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 100; i++ {
		waiter.Add(1)
		listener := strobe.Listener(ctx)
		go func(t *testing.T, waiter *sync.WaitGroup, listener <-chan string) {
			message := <-listener
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
			message := <-stb.Listener(ctx)
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter, strobe)
	}

	for i := 0; i < 2; i++ {
		waiter.Add(1)
		go func(t *testing.T, waiter *sync.WaitGroup) {
			message := <-strobe.Listener(ctx)
			if message == "PULSE" {
				waiter.Done()
			}
		}(t, waiter)
	}

	forgetCtx, forgetCancel := context.WithCancel(context.Background())
	forgottenListener := strobe.Listener(forgetCtx)
	forgetCancel()
	go func() {
		message := <-forgottenListener
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

func Example() {
	s := New[string]()
	w := &sync.WaitGroup{}
	w.Add(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(listener <-chan string) {
		message := <-listener
		fmt.Println(message)
		w.Done()
	}(s.Listener(ctx))
	go func(listener <-chan string) {
		message := <-listener
		fmt.Println(message)
		w.Done()
	}(s.Listener(ctx))
	go func(listener <-chan string) {
		message := <-listener
		fmt.Println(message)
		w.Done()
	}(s.Listener(ctx))

	s.Pulse("PING")
	w.Wait()

	// Output:
	// PING
	// PING
	// PING
}
