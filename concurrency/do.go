package concurrency

import (
	"sync"
)

// DoConcurrently distributes an array of work across a specified number of goroutines
func DoConcurrently[T any](c int, work []T, workFn func(T) error) error {
	requests := make(chan T)
	errors := make(chan error)

	var wg sync.WaitGroup
	for range c {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				r, ok := <-requests
				if !ok {
					return
				}

				err := workFn(r)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	for _, item := range work {
		requests <- item
	}
	close(requests)
	wg.Wait()

	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}
