// This example demonstrates a priority queue built using the heap interface.
package main

import (
	"container/heap"
	"strconv"

	"golang.org/x/exp/slices"
)

// A PriorityQueue implements heap.Interface and holds Items.

type PriorityQueue []*ISISMessage

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	priorityI := pq[i]
	priorityJ := pq[j]
	return !ComparePriorities(priorityI.priority, priorityJ.priority)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*ISISMessage)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Peek() any {
	if len(*pq) == 0 {
		return nil
	}
	return (*pq)[len(*pq)-1]
}

func (pq *PriorityQueue) Print() string {
	contents := "["
	for _, item := range *pq {
		contents += "(" + strconv.FormatBool(item.undeliverable) + ") " + item.Transaction.Identifier + ": " + strconv.Itoa(item.priority.Num) + "." + item.priority.Identifier
		contents += "  "
	}
	contents += "]"
	return contents
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(isis *ISISMessage, priority ISISPriority) {
	// item.priority = priority

	//get item index from queue using identifier
	itemIdx := slices.IndexFunc((*pq), func(m *ISISMessage) bool { return m.Transaction.Identifier == isis.Transaction.Identifier })
	if itemIdx < 0 {
		panic("ITEM NOT IN PQ")
	}

	item := (*pq)[itemIdx]
	item.priority = priority
	item.undeliverable = false // im hoping update is only called when a final priority is achieved, in which case it is now deliverable

	heap.Fix(pq, itemIdx)
}
