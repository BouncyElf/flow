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
	New().With(
		func() {
			data = append(data, 1)
		},
	).With(
		func() {
			data = append(data, 2)
		},
	).With(
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
