package flow

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockPool struct {
	submitErr error
	called    int
}

func (m *mockPool) Submit(fn func()) error {
	if m.submitErr != nil {
		return m.submitErr
	}
	m.called++
	go fn()
	return nil
}

func TestRunner_AddAndWait_Success(t *testing.T) {
	var count int32
	p := &mockPool{}
	r := NewRunner(p)
	for i := 0; i < 10; i++ {
		err := r.Add(func() {
			atomic.AddInt32(&count, 1)
		})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	}
	r.Wait()
	if count != 10 {
		t.Fatalf("expected count 10 got %d", count)
	}
	if p.called != 10 {
		t.Fatalf("expected called 10 got %d", p.called)
	}
}

func TestRunner_Add_FailSubmit(t *testing.T) {
	p := &mockPool{submitErr: errors.New("fail")}
	r := NewRunner(p)
	err := r.Add(func() {})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRunner_Add_NilRunnerOrPool(t *testing.T) {
	var r *Runner = nil
	err := r.Add(func() {})
	if err != ErrInvalidRunner {
		t.Fatalf("expected ErrInvalidRunner, got %v", err)
	}

	r2 := &Runner{}
	err = r2.Add(func() {})
	if err != ErrInvalidRunner {
		t.Fatalf("expected ErrInvalidRunner, got %v", err)
	}
}

func TestRunner_Wait_Empty(t *testing.T) {
	// Wait should return immediately if no Add called
	r := NewRunner(&mockPool{})
	r.Wait()
}

func TestRunner_Add_PanicInFunc(t *testing.T) {
	// Test if panic in f不会引起waitgroup异常
	p := &mockPool{}
	r := NewRunner(p, WithPanicHandler(
		func(msg interface{}) {
			fmt.Println(msg)
		},
	))
	err := r.Add(func() { panic("test") })
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	r.Wait()
}

func TestRunnerWithMultipleEntry(t *testing.T) {
	r, err := NewRunnerWithAntsPool(4)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	var count int32

	f := func() {
		atomic.AddInt32(&count, 1)
	}

	for range make([]struct{}, 10) {
		r.Add(f)
	}
	r.Wait()
	assert.Equal(t, count, int32(10))

	for range make([]struct{}, 10) {
		r.Add(f)
	}
	r.Wait()
	assert.Equal(t, count, int32(20))

	for range make([]struct{}, 10) {
		r.Add(f)
	}
	r.Wait()
	assert.Equal(t, count, int32(30))
}
