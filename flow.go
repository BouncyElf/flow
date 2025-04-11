package flow

import (
	"fmt"
	"sync"
)

// Silent disable all error message
var Silent = false

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
		if f.job_count == 0 {
			return
		}
		taskCh := make(chan func())
		// use min(limit, job_count) to prevent idle worker
		worker_number := f.limit
		if f.job_count < worker_number {
			worker_number = f.job_count
		}
		for range make([]any, worker_number) {
			go func(taskCh chan func()) {
				for job := range taskCh {
					if job == nil {
						taskCh <- nil
						return
					}
					job()
				}
			}(taskCh)
		}
		for _, jobs := range f.jobs {
			wg := new(sync.WaitGroup)
			for _, job := range jobs {
				j := job
				wg.Add(1)
				taskCh <- func() {
					defer func() {
						if msg := recover(); msg != nil {
							f.panicHandler(msg)
						}
						wg.Done()
					}()

					j()
				}
			}
			wg.Wait()
		}
		taskCh <- nil
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
