package flow

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExample(t *testing.T) {
	f, counter := New(), 0
	showLevel := func() {
		counter++
		fmt.Println("level", counter)
	}
	// Next start a new level, and put the func `showLevel` in this level
	// the first level will be created by Flow
	// so you can also use With to add showLevel in the first level
	// f.With(showLevel)
	f.Next(showLevel)
	for i := 0; i < 20; i++ {
		v := i
		// With add func in this level
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
	// wait and add counter
	f.Run()
}

func TestConcurrent(t *testing.T) {
	f, a := New(), 0
	for i := 0; i < 10000; i++ {
		f.With(func() {
			a += 1
		})
	}
	f.Run()
	assert.True(t, a < 10000)
}

func TestOrder(t *testing.T) {
	data := make([]int, 0, 20)
	New().Next(
		func() {
			data = append(data, 1)
		},
	).Next(
		func() {
			data = append(data, 2)
		},
	).Next(
		func() {
			data = append(data, 3)
		},
	).Run()
	assert.Equal(t, []int{1, 2, 3}, data)
}

func TestPanic(t *testing.T) {
	counter, mu := 0, new(sync.Mutex)
	New().With(
		func() {
			panic("i")
		}, func() {
			panic("just")
		}, func() {
			panic("wanna")
		}, func() {
			panic("panic")
		},
	).OnPanic(func(interface{}) {
		mu.Lock()
		defer mu.Unlock()

		counter++
	}).Run()
}

func TestForRangeIssue(t *testing.T) {
	f, a := New(), 0
	for i := 0; i < 10; i++ {
		f.Next(func() {
			a += i
		})
	}
	f.Run()
	assert.True(t, a > 45)

	f, a = New(), 0
	for i := 0; i < 10; i++ {
		v := i
		f.Next(func() {
			a += v
		})
	}
	f.Run()
	assert.Equal(t, 45, a)
}

func TestLimit(t *testing.T) {
	f := New().Limit(4)
	for i := 0; i < 20; i++ {
		f.With(func() {
			assert.True(t, len(f.current) < 5)
		}, func() {
			assert.True(t, len(f.current) < 5)
		}, func() {
			assert.True(t, len(f.current) < 5)
		}, func() {
			assert.True(t, len(f.current) < 5)
		}, func() {
			assert.True(t, len(f.current) < 5)
		}, func() {
			assert.True(t, len(f.current) < 5)
		}, func() {
			assert.True(t, len(f.current) < 5)
		}).Next(func() {
			assert.True(t, len(f.current) == 1)
		}).Next()
	}
	f.Run()
}
