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

type Flow struct {
	nodes        []*node
	errorHandler func(error)
	panicHandler func(interface{})
}

func (f *Flow) reset() *Flow {
	f.nodes = nil
	f.errorHandler = nil
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

func New() *Flow {
	return getFlow()
}

func (f *Flow) With(jobs ...func()) *Flow {
	n := getNode()
	if len(jobs) != 0 {
		for i := 0; i < len(jobs); i++ {
			n.jobs = append(n.jobs, job(jobs[i]))
		}
	}
	f.nodes = append(f.nodes, n)
	return f
}

func (f *Flow) OnError(errorHandler func(error)) *Flow {
	f.errorHandler = errorHandler
	return f
}

func (f *Flow) OnPanic(panicHandler func(interface{})) *Flow {
	f.panicHandler = panicHandler
	return f
}

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
