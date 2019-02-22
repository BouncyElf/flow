package flow

import (
	"fmt"
	"sync"
)

// SilentMode disable all error message
var SilentMode = false

var defaultPanicHandler = func(msg interface{}) {
	say(msg, "panic")
}

type job func()

func (j job) run() {
	if j == nil {
		say("nil job", "error")
		return
	}
	j()
}

type node struct {
	jobs []job
}

func (n *node) reset() *node {
	n.jobs = nil
	return n
}

// Flow is a sync model
type Flow struct {
	nodes        []*node
	panicHandler func(interface{})

	// concurrent limit number
	limit   int
	current chan struct{}

	isNew bool
}

func (f *Flow) reset() *Flow {
	f.isNew = true
	f.nodes = f.nodes[:0]
	f.panicHandler = nil
	f.limit = 0
	f.current = nil
	return f
}

var nodePool *sync.Pool
var flowPool *sync.Pool

func init() {
	nodePool = new(sync.Pool)
	nodePool.New = func() interface{} {
		return newNode()
	}
	flowPool = new(sync.Pool)
	flowPool.New = func() interface{} {
		return newFlow()
	}
}

// New returns a flow instance
func New() *Flow {
	return getFlow()
}

// With add funcs in this level
// With: run f1, run f2, run f3 ... (random execute order)
func (f *Flow) With(jobs ...func()) *Flow {
	if len(f.nodes) == 0 {
		f.nodes = append(f.nodes, getNode())
	}
	n := f.nodes[len(f.nodes)-1]
	for i := 0; i < len(jobs); i++ {
		n.jobs = append(n.jobs, job(jobs[i]))
	}
	return f
}

// Next add funcs in next level
// Next: wait level1(run f1, run f2, run f3...) ... wait level2(...)... (in order)
func (f *Flow) Next(jobs ...func()) *Flow {
	f.nodes = append(f.nodes, getNode())
	f.With(jobs...)
	return f
}

// OnPanic set panicHandler
func (f *Flow) OnPanic(panicHandler func(interface{})) *Flow {
	f.panicHandler = panicHandler
	return f
}

// Limit limit the number of concurrent goroutines
func (f *Flow) Limit(number int) *Flow {
	if number <= 0 {
		return f
	}
	f.limit = number
	f.current = make(chan struct{}, number)
	return f
}

// Run execute these funcs
func (f *Flow) Run() {
	if f == nil || !f.isNew {
		say("invalid flow", "error")
		return
	}
	panicHandler := defaultPanicHandler
	if f.panicHandler != nil {
		panicHandler = f.panicHandler
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < len(f.nodes); i++ {
		for j := 0; j < len(f.nodes[i].jobs); j++ {
			if f.limit > 0 {
				f.current <- struct{}{}
			}
			wg.Add(1)
			go func(i, j int) {
				defer func() {
					if msg := recover(); msg != nil {
						panicHandler(msg)
					}
					if f.limit > 0 {
						<-f.current
					}
					wg.Done()
				}()
				f.nodes[i].jobs[j].run()
			}(i, j)
		}
		nodePool.Put(f.nodes[i])
		wg.Wait()
	}
	flowPool.Put(f)
	f.isNew = false
}

func say(msg interface{}, level string) {
	if !SilentMode {
		fmt.Printf("%s %s: %v\n", "flow", level, msg)
	}
}

func getNode() *node {
	return nodePool.Get().(*node).reset()
}

func getFlow() *Flow {
	return flowPool.Get().(*Flow).reset()
}

func newNode() *node {
	return new(node)
}

func newFlow() *Flow {
	return new(Flow)
}
