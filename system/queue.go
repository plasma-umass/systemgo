package system

import (
	"container/heap"
	"reflect"

	"github.com/b1101/systemgo/lib/handle"
)

type Queue struct {
	queue      []*Unit
	queued     map[*Unit]bool
	add, Start chan *Unit
}

func NewQueue() (q *Queue) {
	q = &Queue{
		[]*Unit{},
		map[*Unit]bool{},
		make(chan *Unit),
		make(chan *Unit),
	}
	heap.Init(q)
	go func() {
		for {
			select {
			case u := <-q.add:
				heap.Push(q, u)
			default:
				for q.Len() > 0 {
					q.Start <- heap.Pop(q).(*Unit)
				}
				heap.Push(q, <-q.add)
			}
		}
	}()
	return
}

func (q Queue) Len() int {
	return len(q.queue)
}

func (q Queue) Less(i, j int) bool {
	for _, name := range q.queue[j].Requires() {
		if q.queue[i].Name() == name {
			return true
		}
	}
	for _, name := range q.queue[j].Wants() {
		if q.queue[i].Name() == name {
			return true
		}
	}
	for _, name := range q.queue[j].After() {
		if q.queue[i].Name() == name {
			return true
		}
	}
	return false
}

func (q *Queue) Swap(i, j int) {
	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
}

func (q *Queue) Push(x interface{}) {
	if u, ok := x.(*Unit); !ok {
		handle.Serr("Could not assert element to *system.Unit\n", reflect.TypeOf(x))
	} else {
		if !q.queued[u] {
			q.queue = append(q.queue, u)
			q.queued[u] = true
		}
	}
}

func (q *Queue) Pop() interface{} {
	u := q.queue[q.Len()-1]
	delete(q.queued, u)
	q.queue = q.queue[:q.Len()-1]
	return u
}

func (q *Queue) Add(u *Unit) {
	q.add <- u
}

func (q *Queue) Remove(u *Unit) {
	if q.queued[u] {
		for i, enqueued := range q.queue {
			if enqueued == u {
				heap.Remove(q, i)
				heap.Fix(q, i)
				delete(q.queued, u)
				break
			}
		}
	}
}
