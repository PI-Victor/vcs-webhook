package queue

import (
	"github.com/vastness-io/queues/pkg/core"
	"sync"
)

// fifoQueue is a first-In-First-Out implementation of Queue.
type fifoQueue struct {
	head     int // head is the first index of the slice.
	nodes    []interface{}
	cond     *sync.Cond
	count    int
	tail     int  // tail is the last index of the slice.
	shutdown bool // Is the queue been signalled to shutdown.
}

// NewFIFOQueue creates a new First-In-First-Out Queue.
func NewFIFOQueue() core.BlockingQueue {
	return &fifoQueue{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

// Size of the queue.
func (q *fifoQueue) Size() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return q.count
}

// Enqueue adds the node to the queue.
func (q *fifoQueue) Enqueue(node interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.nodes = append(q.nodes, node)
	q.count++
	q.tail++
	q.cond.Signal()
}

// Dequeue removes and returns the node from the Head of the queue.
func (q *fifoQueue) Dequeue() (interface{}, bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.count == 0 && !q.shutdown {
		q.cond.Wait()
	}

	if q.count == 0 {
		//Shutting down
		return nil, true
	}

	head := q.nodes[q.head]
	q.nodes = q.nodes[1:]
	q.count--
	q.tail--
	return head, false
}

// Shutdown signals the queue to shutdown. This is necessary to notify go routines which are currently blocked on trying to dequeue.
func (q *fifoQueue) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shutdown = true
	q.cond.Broadcast()
}
