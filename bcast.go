// Package bcast provides a simple N:M channel-based broadcaster.
package bcast

import (
	"sync"
)

// Broadcast is an N:M channel-based broadcaster. It allows for
// event-like broadcasts of arbitrary data to multiple receivers
// simultaneously. The order that data is sent in is guaranteed to be
// received in that order by any listeners, but the order that the
// listeners receive the data in relative to each other is not.
//
// The zero-value is valid and ready for use, but clients should make
// sure that they don't use multiple copies of the same instance.
type Broadcast struct {
	initOnce sync.Once

	cancel sync.Once
	done   chan struct{}

	listen    chan listener
	stop      chan chan<- interface{}
	broadcast chan interface{}
}

// init does lazy initialization of the Broadcast.
func (bc *Broadcast) init() {
	bc.initOnce.Do(func() {
		bc.done = make(chan struct{})

		bc.listen = make(chan listener)
		bc.stop = make(chan chan<- interface{})
		bc.broadcast = make(chan interface{})

		go bc.coord()
	})
}

// coord runs the main coordinator of the broadcasting system. It
// registers new listeners, deletes and closes listeners, and
// broadcasts data to any existing listeners.
func (bc *Broadcast) coord() {
	listeners := make(map[chan<- interface{}]<-chan struct{})
	defer func() {
		for c := range listeners {
			close(c)
		}
	}()

	for {
		select {
		case <-bc.done:
			return

		case c := <-bc.listen:
			listeners[c.out] = c.done

		case c := <-bc.stop:
			if _, ok := listeners[c]; ok {
				delete(listeners, c)
				close(c)
			}

		case data := <-bc.broadcast:
			for c, stop := range listeners {
				select {
				case c <- data:
				case <-stop:
				}
			}
		}
	}
}

// Listen registers c as a listening channel, meaning that any
// broadcasts sent after this call will be replicated to c. The
// returned function unregisters c.
//
// When a listening channel is done, either because the returned stop
// function was called or because the entire broadcaster was stopped,
// it is closed.
//
// This function will block if an existing broadcast is in progress.
//
// This function is a no-op if the broadcaster has been stopped, as is
// calling the returned stop function.
func (bc *Broadcast) Listen(c chan<- interface{}) (stop func()) {
	bc.init()

	done := make(chan struct{})
	lis := listener{
		out:  c,
		done: done,
	}

	select {
	case <-bc.done:
	case bc.listen <- lis:
	}

	var cancel sync.Once

	return func() {
		cancel.Do(func() {
			close(done)

			select {
			case <-bc.done:
			case bc.stop <- c:
			}
		})
	}
}

// Send returns a channel which broadcasts anything sent to it to all
// listening channels. Sending to this channel will block if any
// existing broadcasts are in progress. A broadcast is considered to
// be in progress until all listening channels have been sent to
// successfully.
//
// It is the callers responsibility to not send anything to the
// channel returned by Send after the broadcaster has been stopped.
func (bc *Broadcast) Send() chan<- interface{} {
	bc.init()

	return bc.broadcast
}

// Stop stops the broadcaster, resulting in all listening channels
// being closed and the background coordinator being stopped.
func (bc *Broadcast) Stop() {
	bc.init()

	bc.cancel.Do(func() {
		close(bc.done)
	})
}

type listener struct {
	out  chan<- interface{}
	done <-chan struct{}
}
