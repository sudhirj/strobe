# strobe 

[![CI](https://github.com/RealImage/WireApp/actions/workflows/ci.yaml/badge.svg)](https://github.com/RealImage/WireApp/actions/workflows/ci.yaml) [![Documentation](https://godoc.org/github.com/sudhirj/strobe?status.svg)](https://godoc.org/github.com/sudhirj/strobe)

Go channel fan-out. Send the same message to many channels simultaneously.

## Example

```go
func Example() {
	s := NewStrobe[string]()
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
```

