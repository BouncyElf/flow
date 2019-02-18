package flow

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrent(t *testing.T) {
	New().With(
		func() {
			fmt.Println(1)
		}, func() {
			fmt.Println(2)
		}, func() {
			fmt.Println(3)
		},
	).Run()
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
	).Run()
}

func TestForRangeIssue(t *testing.T) {
	f, a := New(), 0
	for i := 0; i < 10; i++ {
		f.Next(func() {
			a += i
		})
	}
	f.Run()
	assert.Equal(t, 100, a)
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
