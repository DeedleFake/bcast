package bcast

import (
	"sync"
)

// Broadcast is a N:M channel-based broadcaster. It allows for
// event-like broadcasts of arbitrary data to multiple receivers
// simultaneously, with guarantees on ordering and receipt.
//
// The zero-value is valid and ready for use, but clients should make
// sure that they don't use multiple copies of the same instance.
type Broadcast struct {
	initOnce sync.Once

	cancel sync.Once
	done   chan struct{}

	listen    chan chan<- interface{}
	stop      chan chan<- interface{}
	broadcast chan interface{}
}

func (bc *Broadcast) init() {
	bc.initOnce.Do(func() {
		bc.done = make(chan struct{})

		bc.listen = make(chan chan<- interface{})
		bc.stop = make(chan chan<- interface{})
		bc.broadcast = make(chan interface{})

		go bc.coord()
	})
}

func (bc *Broadcast) coord() {
	listeners := make(map[chan<- interface{}]struct{})
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
			listeners[c] = struct{}{}

		case c := <-bc.stop:
			if _, ok := listeners[c]; ok {
				delete(listeners, c)
				close(c)
			}

		case data := <-bc.broadcast:
			for c := range listeners {
				c <- data
			}
		}
	}
}

func (bc *Broadcast) Listen(c chan<- interface{}) (stop func()) {
	bc.init()

	bc.listen <- c
	return func() {
		select {
		case <-bc.done:
		case bc.stop <- c:
		}
	}
}

func (bc *Broadcast) Send() chan<- interface{} {
	bc.init()

	return bc.broadcast
}

func (bc *Broadcast) Stop() {
	bc.init()

	bc.cancel.Do(func() {
		close(bc.done)
	})
}
