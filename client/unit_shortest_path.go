package main

import (
	"container/heap"
	"fmt"
)

// shortestPathToFinish will find the shortest path for u to reach the end of the map avoiding
// all the Towers (tws) and only moving on the map valid area.
// It returns the paths to follow in order, path[0] is the next move and path[-1] is the last one
// there are no DIAGONAL MOVES just UP, DOWN, LEFT and RIGHT
func (us *UnitsStore) shortestPathToFinish(m *Map, lid int, u MovingEntity, tws []Object) []Step {
	visited := make(map[string]struct{})
	dists := make(map[string]float64)
	prev := make(map[Step]Step)

	start := Step{
		Object: Object{
			Y: u.Y,
			X: u.X,
			// W and H need to be 1 or it can pass in between the towers
			W: 1, H: 1,
		},
		Facing: u.Facing,
	}

	dists[calculateObjectKey(start.Object)] = 0
	queue := &queue{&queueItem{value: start, weight: 0, index: 0}}
	heap.Init(queue)

	var endPosition Step
	for queue.Len() > 0 {
		item := heap.Pop(queue).(*queueItem)
		s := item.value

		if _, ok := visited[calculateObjectKey(s.Object)]; ok {
			continue
		}

		visited[calculateObjectKey(s.Object)] = struct{}{}

		// If the new position will make it hit a tower
		// we have to skip that move
		if isCollidingWithTower(s.Object, tws) {
			continue
		}

		// We want to validate first if the unit is at the end
		// before we validate it's on a valid UnitZone as TheEnd
		// is not a valid UnitZone
		if m.IsAtTheEnd(s.Object, lid) {
			endPosition = s
			break
		}

		// We validate that when the 'u' moves to the next place
		// it does not end up in an invalid zone
		if !m.IsInValidUnitZone(s.Object, lid) {
			continue
		}

		for _, ns := range s.NeighborSteps() {
			dist := dists[calculateObjectKey(s.Object)] + 1
			if tentativeDist, ok := dists[calculateObjectKey(ns.Object)]; !ok || dist < tentativeDist {
				dists[calculateObjectKey(ns.Object)] = dist
				prev[ns] = s
				heap.Push(queue, &queueItem{value: ns, weight: dist})
			}
		}

	}

	path := []Step{endPosition}
	for next := prev[endPosition]; next != start; next = prev[next] {
		path = append(path, next)
	}
	path = append(path, start)

	// Reverse path.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path
}

func calculateObjectKey(o Object) string {
	return fmt.Sprintf("%s%s", o.X, o.Y)
}

type queueItem struct {
	value  Step
	weight float64
	index  int
}

type queue []*queueItem

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].weight < q[j].weight
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue) Push(x interface{}) {
	n := len(*q)
	item := x.(*queueItem)
	item.index = n
	*q = append(*q, item)
}

func (q *queue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*q = old[0 : n-1]
	return item
}

func isCollidingWithTower(o Object, tws []Object) bool {
	for _, t := range tws {
		if o.IsColliding(t) {
			return true
		}
	}
	return false
}
