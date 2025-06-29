package flow

import (
	"errors"
	"sync"

	"github.com/panjf2000/ants"
)

type PanicHandler func(interface{})

// Runner allows you run the functions while adding it into runner.
type Runner struct {
	p  GroutinePool
	wg *sync.WaitGroup
	ph PanicHandler
}

var (
	ErrInvalidRunner = errors.New("invalid runner")
)

type RunnerOptions func(*Runner)

func NewRunnerWithAntsPool(size int, ro ...RunnerOptions) (*Runner, error) {
	p, err := ants.NewPool(size)
	if err != nil {
		return nil, err
	}
	return NewRunner(p, ro...), nil
}

func NewRunner(p GroutinePool, ro ...RunnerOptions) *Runner {
	if p == nil {
		p = globalPool
	}
	r := &Runner{
		p:  p,
		wg: new(sync.WaitGroup),
	}
	for _, o := range ro {
		o(r)
	}
	return r
}

func WithPanicHandler(ph PanicHandler) RunnerOptions {
	return func(r *Runner) {
		r.ph = ph
	}
}

// `Add` add the f into the pool and the pool will run f
func (r *Runner) Add(f func()) error {
	if r == nil || r.p == nil {
		return ErrInvalidRunner
	}
	r.wg.Add(1)
	err := r.p.Submit(
		func() {
			defer r.wg.Done()
			defer func() {
				if r.ph != nil {
					if msg := recover(); msg != nil {
						r.ph(msg)
					}
				}
			}()
			f()
		},
	)
	if err != nil {
		r.wg.Done()
		return err
	}
	return nil
}

// `Wait` waits all the functions inside the pool execute finished.
func (r *Runner) Wait() {
	if r == nil || r.wg == nil {
		return
	}
	r.wg.Wait()
}
