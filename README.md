# Flow
🌊Flow is an easy-used concurrent calculate model in Go.
Flow can easily limit the maxium number of concurrent goroutine and can be very straight-forward to use.
Basically, flow is a 2d matrix and it runs jobs row-by-row and all the jobs in same row will run concurrently.

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
	// Next start a new level, and put the job `showLevel` in this level
	// the first level will be created by Flow
	// so you can also use With to add showLevel in the first level
	// f.With(showLevel)
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
