# Flow
🌊Flow is an easy-used concurrent calculate model in Go.

Flow can easily limit the maxium number of concurrent goroutine and can be very straight-forward to use.

Basically, flow is a 2D matrix and it run jobs row-by-row which in same row, the jobs will run concurrently.

# Example
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
