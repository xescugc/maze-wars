package store

import (
	"container/heap"
	"fmt"

	"github.com/xescugc/ltw/utils"
)

// stepMap is a collection of steps for quick reference
type stepMap map[string]*queueItem

// get gets the Pather object wrapped in a node, instantiating if required.
func (sm stepMap) get(s utils.Step) *queueItem {
	k := calculateObjectKey(s.Object)
	qi, ok := sm[k]
	if !ok {
		qi = &queueItem{
			step: s,
		}
		sm[k] = qi
	}
	return qi
}

// astar will find the shortest path for u to reach the end of the map avoiding
// all the Towers (tws) and only moving on the map valid area.
// It returns the paths to follow in order, path[0] is the next move and path[-1] is the last one
// there are no DIAGONAL MOVES just UP, DOWN, LEFT and RIGHT
func (us *Units) astar(m *Map, lid int, u utils.MovingObject, tws []utils.Object) []utils.Step {
	nm := stepMap{}
	nq := &queue{}
	heap.Init(nq)
	from := utils.Step{
		Object: utils.Object{
			Y: u.Y,
			X: u.X,
			// W and H need to be 1 or it can pass in between the towers
			W: 1, H: 1,
		},
		Facing: u.Facing,
	}
	fromStep := nm.get(from)
	fromStep.open = true
	heap.Push(nq, fromStep)
	for {
		if nq.Len() == 0 {
			// There's no path, return found false.
			return nil
		}
		current := heap.Pop(nq).(*queueItem)
		current.open = false
		current.closed = true

		if m.IsAtTheEnd(current.step.Object, lid) {
			// Found a path to the goal.
			p := []utils.Step{}
			curr := current
			for curr != nil {
				p = append(p, curr.step)
				curr = curr.parent
			}
			for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
				p[i], p[j] = p[j], p[i]
			}
			return p
		}

		// If the new position will make it hit a tower
		// we have to skip that move
		if isCollidingWithTower(current.step.Object, tws) {
			continue
		}

		// We validate that when the 'u' moves to the next place
		// it does not end up in an invalid zone
		if !m.IsInValidUnitZone(current.step.Object, lid) {
			continue
		}

		for _, neighbor := range current.step.NeighborSteps() {
			// The cost to the neighbor is always 1
			cost := current.cost + 1
			neighborStep := nm.get(neighbor)
			if cost < neighborStep.cost {
				if neighborStep.open {
					heap.Remove(nq, neighborStep.index)
				}
				neighborStep.open = false
				neighborStep.closed = false
			}
			if !neighborStep.open && !neighborStep.closed {
				neighborStep.cost = cost
				neighborStep.open = true
				to := straightPathToEnd(m, lid, neighbor)
				neighborStep.rank = cost + neighbor.Distance(to)
				neighborStep.parent = current
				heap.Push(nq, neighborStep)
			}
		}
	}
}

func straightPathToEnd(m *Map, lid int, s utils.Step) utils.Object {
	end := utils.Object{
		X: s.X,
		Y: m.EndZone(lid).Y,
	}
	return end
}

func calculateObjectKey(o utils.Object) string {
	return fmt.Sprintf("%s%s", o.X, o.Y)
}

type queueItem struct {
	step   utils.Step
	cost   float64
	rank   float64
	parent *queueItem
	open   bool
	closed bool
	index  int
}

type queue []*queueItem

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].rank < q[j].rank
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

func isCollidingWithTower(o utils.Object, tws []utils.Object) bool {
	for _, t := range tws {
		if o.IsColliding(t) {
			return true
		}
	}
	return false
}
