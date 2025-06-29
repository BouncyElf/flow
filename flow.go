package flow

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/panjf2000/ants"
)

// Silent disable all error message
var Silent = false

// ants pool (global)
var (
	poolSize     = 100
	globalPool   *ants.Pool
	once         sync.Once
	defaultLimit = runtime.NumCPU()
)

// GoroutinePool is the goroutine pool where you want to execute your job
type GroutinePool interface {
	// Submit add the function into the pool.
	// it should ALWAYS returns an error when it fail to add the function into the pool.
	Submit(func()) error
}

type ErrorHandler func(error)

func init() {
	initGlobalPool()
	if defaultLimit < 1 {
		defaultLimit = 1
	}
}

// initGlobalPool initializes the global ants pool
func initGlobalPool() {
	once.Do(func() {
		p, err := ants.NewPool(poolSize)
		if err != nil {
			panic(fmt.Sprintf("failed to create global ants pool: %v", err))
		}
		globalPool = p
	})
}

// SetGlobalPoolSize sets global ants pool size (must be called before any task runs)
func SetGlobalPoolSize(size int) error {
	if size < 1 {
		size = 1
	}
	poolSize = size
	if globalPool != nil {
		globalPool.Release()
		globalPool = nil
	}
	p, err := ants.NewPool(poolSize)
	if err != nil {
		return err
	}
	globalPool = p
	return nil
}

// globalLimit limit all flow's concurrent work number.
// <= 0 means no limits.
var defaultPanicHandler = func(msg interface{}) {
	say(msg, "panic")
}

// Flow is a sync model
type Flow struct {
	jobs         [][]func()
	job_count    int
	panicHandler func(interface{})
	errorHandler ErrorHandler

	p GroutinePool

	// concurrent limit number
	limit int

	runOnce *sync.Once
}

// New returns a flow instance
func New() *Flow {
	return &Flow{
		jobs:         [][]func(){},
		panicHandler: defaultPanicHandler,
		limit:        defaultLimit,
		runOnce:      new(sync.Once),
	}
}

func NewWithPool(p GroutinePool) *Flow {
	f := New()
	f.SetGroutinePool(p)
	return f
}

func NewWithLimit(limit int) *Flow {
	f := New()
	f.SetLimit(limit)
	return f
}

func (f *Flow) SetErrorHandler(h ErrorHandler) *Flow {
	f.errorHandler = h
	return f
}

func (f *Flow) SetGroutinePool(p GroutinePool) *Flow {
	if p != nil {
		f.p = p
	}
	return f
}

// SetLimit set the max concurrent goroutines number
func (f *Flow) SetLimit(limit int) *Flow {
	if limit < 1 {
		limit = 1
	}
	f.limit = limit
	return f
}

// With add funcs in this level
// With: run f1, run f2, run f3 ... (random execute order)
func (f *Flow) With(jobs ...func()) *Flow {
	if len(f.jobs) == 0 {
		f.jobs = make([][]func(), 1)
	}
	n := len(f.jobs)
	f.jobs[n-1] = append(f.jobs[n-1], jobs...)
	f.job_count += len(jobs)
	return f
}

// Next add funcs in next level
// Next: wait level1(run f1, run f2, run f3...) ... wait level2(...)... (in order)
func (f *Flow) Next(jobs ...func()) *Flow {
	f.jobs = append(f.jobs, jobs)
	f.job_count += len(jobs)
	return f
}

// Run execute these funcs
func (f *Flow) Run() {
	f.runOnce.Do(func() {
		if f.job_count == 0 || len(f.jobs) == 0 {
			return
		}

		for _, jobs := range f.jobs {
			wg := new(sync.WaitGroup)

			sem := make(chan struct{}, f.limit) // Per-level concurrency limit
			for _, job := range jobs {
				j := job
				wg.Add(1)

				sem <- struct{}{} // block if over limit

				p := f.p
				if p == nil {
					p = globalPool
				}
				err := p.Submit(func() {
					defer func() {
						if msg := recover(); msg != nil {
							f.panicHandler(msg)
						}
						<-sem // release slot
						wg.Done()
					}()
					j()
				})
				if err != nil && f.errorHandler != nil {
					f.errorHandler(err)
				}
			}
			wg.Wait()
		}
	})
}

// OnPanic set panicHandler
func (f *Flow) OnPanic(panicHandler func(interface{})) *Flow {
	f.panicHandler = panicHandler
	return f
}

func say(msg interface{}, level string) {
	if Silent {
		return
	}
	fmt.Printf("%s %s: %v\n", "flow", level, msg)
}
