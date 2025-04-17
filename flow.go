package flow

import (
	"fmt"
	"sync"

	"github.com/panjf2000/ants"
)

// Silent disable all error message
var Silent = false

// ants pool (global)
var (
	poolSize   = 100
	globalPool *ants.Pool
	once       sync.Once
)

func init() {
	initGlobalPool()
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

	// concurrent limit number
	limit int

	runOnce *sync.Once
}

// New returns a flow instance
func New() *Flow {
	return &Flow{
		jobs:         [][]func(){},
		panicHandler: defaultPanicHandler,
		limit:        10,
		runOnce:      new(sync.Once),
	}
}

func NewWithLimit(limit int) *Flow {
	f := New()
	f.SetLimit(limit)
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

				_ = globalPool.Submit(func() {
					defer func() {
						if msg := recover(); msg != nil {
							f.panicHandler(msg)
						}
						<-sem // release slot
						wg.Done()
					}()
					j()
				})
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
