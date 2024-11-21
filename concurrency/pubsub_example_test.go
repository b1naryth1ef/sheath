package concurrency_test

import (
	"fmt"
	"time"

	"github.com/b1naryth1ef/sheath/concurrency"
)

func ExamplePubSub() {
	pubsub := concurrency.NewPubSub[int]()

	results := make(chan int, 32)

	for range 4 {
		go func() {
			messages, done := pubsub.Subscribe()
			defer done()

			for {
				item, ok := <-messages
				if !ok {
					return
				}
				results <- item
			}
		}()
	}

	time.Sleep(time.Millisecond)
	pubsub.Publish(0)

	time.Sleep(time.Millisecond)
	pubsub.Publish(1)

	time.Sleep(time.Millisecond)
	pubsub.Publish(2)

	values := []int{}
	for range 12 {
		item := <-results
		values = append(values, item)
	}

	fmt.Printf("%v\n", values)
	// Output: [0 0 0 0 1 1 1 1 2 2 2 2]
}
