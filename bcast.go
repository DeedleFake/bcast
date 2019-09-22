package bcast

import (
	"context"
)

type Broadcast struct {
	cancel context.CancelFunc

	listen    chan listener
	stop      chan chan<- interface{}
	broadcast chan interface{}
}

func New() *Broadcast {
	return WithContext(context.Background())
}

func WithContext(ctx context.Context) *Broadcast {
	ctx, cancel := context.WithCancel(ctx)

	bc := &Broadcast{
		cancel: cancel,

		listen:    make(chan listener),
		stop:      make(chan chan<- interface{}),
		broadcast: make(chan interface{}),
	}
	go bc.coord(ctx)

	return bc
}

func (bc *Broadcast) coord(ctx context.Context) {
	listeners := make(map[chan<- interface{}]context.CancelFunc)

	for {
		select {
		case <-ctx.Done():
			return

		case listener := <-bc.listen:
			ctx, cancel := context.WithCancel(ctx)
			listeners[listener.c] = cancel
			listener.ctx <- ctx

		case listener := <-bc.stop:
			if cancel, ok := listeners[listener]; ok {
				delete(listeners, listener)
				cancel()
			}

		case data := <-bc.broadcast:
			for listener := range listeners {
				listener <- data
			}
		}
	}
}

func (bc *Broadcast) Listen(c chan<- interface{}) (stop func()) {
	ctx := make(chan<- context.Context, 1)

	bc.listen <- listener{
		c:   c,
		ctx: ctx,
	}

	return func() {
		bc.stop <- c
	}
}

func (bc *Broadcast) Send() chan<- interface{} {
	return bc.broadcast
}

func (bc *Broadcast) Stop() {
	bc.cancel()
}

type listener struct {
	c   chan<- interface{}
	ctx chan<- context.Context
}
