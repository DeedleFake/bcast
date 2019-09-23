package bcast_test

import (
	"sync"
	"testing"

	"github.com/DeedleFake/bcast"
)

func TestBroadcast(t *testing.T) {
	tests := []struct {
		name string
		data []interface{}
	}{
		{
			name: "Simple",
			data: []interface{}{"this", "is", "a", "test"},
		},
	}

	for _, test := range tests {
		var bc bcast.Broadcast

		var wg sync.WaitGroup
		for i := 0; i < 3; i++ {
			c := make(chan interface{})
			bc.Listen(c)

			wg.Add(1)
			go func(n int) {
				defer wg.Done()

				var i int
				for data := range c {
					if data != test.data[i] {
						t.Errorf("%v: data[%v] is %q, not %q", n, i, data, test.data[i])
					}

					i++
				}
			}(i)
		}

		send := bc.Send()
		for _, data := range test.data {
			send <- data
		}
		bc.Stop()

		wg.Wait()
	}
}
