# Flow
🌊Flow is an easy-used concurrent calculation model to handle CPU-intensive jobs in Go.

Flow can easily limit concurrent job count and can be very straight-forward to use.

Flow use [ants](https://github.com/panjf2000/ants) to reuse goroutine.

Basically, flow is a 2D matrix and it run jobs row-by-row and in same row, the jobs will run concurrently.

# Usage
Minimal Go Version is 1.18

```Go
package main

import (
	"fmt"

	"github.com/BouncyElf/flow"
)

func main() {
	f, counter := flow.New(), 0
	// declare a job
	showLevel := func() {
		counter++
		fmt.Println("level", counter)
	}
	// Next starts a new level, and put the job `showLevel` in it
	f.Next(showLevel)
	for i := 0; i < 20; i++ {
		v := i
		// With add job in this level
		f.With(func() {
			// do some stuff
			fmt.Println(v)
		})
	}
	// start a new level
	f.Next(showLevel)
	for i := 0; i < 20; i++ {
		v := i
		f.With(func() {
			// do some stuff
			fmt.Println(v)
		})
	}
	// limit the number of concurrent jobs
	f.Limit(10)
	// wait and execute job
	f.Run()
}
```

# Doc
[GoDoc](https://godoc.org/github.com/BouncyElf/flow)
