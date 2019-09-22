package bcast_test

import (
	"sync"
	"testing"

	"github.com/DeedleFake/bcast"
)

func TestBroadcast(t *testing.T) {
	var bc bcast.Broadcast

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		c := make(chan interface{})
		bc.Listen(c)

		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for data := range c {
				t.Logf("%v: %v", i, data)
			}
		}(i)
	}

	bc.Send() <- "this"
	bc.Send() <- "is"
	bc.Send() <- "a"
	bc.Send() <- "test"
	bc.Stop()

	wg.Wait()
}
