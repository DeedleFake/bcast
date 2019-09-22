package bcast

import (
	"sync"
)

type Broadcast struct {
	cancel sync.Once
	done   chan struct{}

	listen    chan chan<- interface{}
	stop      chan chan<- interface{}
	broadcast chan interface{}
}

func New() *Broadcast {
	bc := &Broadcast{
		done: make(chan struct{}),

		listen:    make(chan chan<- interface{}),
		stop:      make(chan chan<- interface{}),
		broadcast: make(chan interface{}),
	}
	go bc.coord()

	return bc
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
	bc.listen <- c
	return func() {
		select {
		case <-bc.done:
		case bc.stop <- c:
		}
	}
}

func (bc *Broadcast) Send() chan<- interface{} {
	return bc.broadcast
}

func (bc *Broadcast) Stop() {
	bc.cancel.Do(func() {
		close(bc.done)
	})
}
