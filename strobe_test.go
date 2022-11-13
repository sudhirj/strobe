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

func TestListenerRaces(t *testing.T) {
	const n = 1000

	sb := New[struct{}]()
	cancels := make([]context.CancelFunc, 0, n)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create n listeners
	for i := 0; i < n; i++ {
		lCtx, lCancel := context.WithCancel(ctx)
		sb.Listener(lCtx)
		cancels = append(cancels, lCancel)
	}

	// race listener ctx cancel with Pulse & Count
	go func() {
		for i := 0; i < 10000; i++ {
			sb.Pulse(struct{}{})
			sb.Count()
		}
	}()
	go func() {
		for _, cancel := range cancels {
			cancel()
		}
	}()
}

func ExampleStrobe() {
	sb := New[string]()
	w := &sync.WaitGroup{}

	listenPrinter := func(l <-chan string) {
		fmt.Println(<-l)
		w.Done()
	}

	w.Add(3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go listenPrinter(sb.Listener(ctx))
	go listenPrinter(sb.Listener(ctx))
	go listenPrinter(sb.Listener(ctx))

	sb.Pulse("PING")
	w.Wait()

	// Output:
	// PING
	// PING
	// PING
}
