package concurrency_test

import (
	"fmt"
	"sync/atomic"

	"github.com/b1naryth1ef/sheath/concurrency"
)

func ExampleDoConcurrently() {
	work := make([]int, 10000)

	var totalSync int64
	concurrency.DoConcurrently(12, work, func(i int) error {
		atomic.AddInt64(&totalSync, 1)
		return nil
	})

	fmt.Printf("%d\n", totalSync)
	// Output: 10000
}
