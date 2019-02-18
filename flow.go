package flow

import (
	"fmt"
	"sync"
)

type job func()

func (j job) run() {
	j()
}

type node struct {
	jobs []job
}

var defaultPanicHandler = func(msg interface{}) {
	fmt.Printf("%s panic with: %v\n", "flow", msg)
}

func (n *node) reset() *node {
	n.jobs = nil
	return n
}

// Flow is a sync model
type Flow struct {
	nodes        []*node
	panicHandler func(interface{})
}

func (f *Flow) reset() *Flow {
	f.nodes = nil
	f.panicHandler = nil
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
	f.nodes[len(f.nodes)-1] = n
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

// Run execute these funcs
func (f *Flow) Run() {
	panicHandler := defaultPanicHandler
	if f.panicHandler != nil {
		panicHandler = f.panicHandler
	}
	wg := new(sync.WaitGroup)
	for i := 0; i < len(f.nodes); i++ {
		for j := 0; j < len(f.nodes[i].jobs); j++ {
			wg.Add(1)
			go func(i, j int) {
				defer func() {
					if msg := recover(); msg != nil {
						panicHandler(msg)
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
