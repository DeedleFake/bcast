package bcast_test

import (
	"sync"
	"testing"

	"github.com/DeedleFake/bcast"
)

func TestBroadcast(t *testing.T) {
	bc := bcast.New()

	tester := func(n int) {
		c := make(chan interface{})
		stop := bc.Listen(c)
		defer stop()

		for data := range c {
			t.Logf("%v: %v", n, data)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tester(i)
		}(i)
	}

	bc.Send() <- "this"
	bc.Send() <- "is"
	bc.Send() <- "a"
	bc.Send() <- "test"
	bc.Stop()

	wg.Wait()
}
