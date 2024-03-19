package graph

import (
	"container/heap"

	"github.com/xescugc/maze-wars/utils"
)

// stepMap is a collection of steps for quick reference
type stepMap map[string]*queueItem

// get gets the Pather object wrapped in a node, instantiating if required.
func (sm stepMap) get(s Step) *queueItem {
	qi, ok := sm[s.Node.ID]
	if !ok {
		qi = &queueItem{
			step: s,
		}
		sm[s.Node.ID] = qi
	}
	return qi
}

type queueItem struct {
	step Step

	parent *queueItem

	cost   int
	rank   int
	open   bool
	closed bool
	index  int
}

// AStar calculates the shortest path between between Source(sx,sy)
// to Target(tx,ty) with W,H equal to the Scale
// If atScale is true it'll return the 1:1 result, if not it'll return
// the 1:Scale result
func (g *Graph) AStar(sx, sy int, d utils.Direction, tx, ty int, atScale bool) []Step {
	nm := stepMap{}
	nq := &queue{}
	heap.Init(nq)
	var (
		sn, tn *Node
	)
	if atScale {
		sn = g.GetNodeOf(sx, sy)
		tn = g.GetNodeOf(tx, ty)
	} else {
		sn = g.GetNode(sx, sy)
		tn = g.GetNode(tx, ty)
	}

	if sn == nil || tn == nil {
		return nil
	}
	ss := Step{
		Node:   sn,
		Facing: d,
	}
	sqi := nm.get(ss)
	sqi.open = true
	heap.Push(nq, sqi)
	for {
		if nq.Len() == 0 {
			// There's no path, return found false.
			return nil
		}
		current := heap.Pop(nq).(*queueItem)
		current.open = false
		current.closed = true

		if current.step.Node.ID == tn.ID || checkConsecutiveSteps(current, 4) {
			if current.step.Node.NextStep != nil {
				current = &queueItem{
					step:   *current.step.Node.NextStep,
					parent: current,
				}
				for current.step.Node.NextStep != nil {
					current = &queueItem{
						step:   *current.step.Node.NextStep,
						parent: current,
					}
				}
			}
			// Found a path to the goal.
			p := []Step{}
			curr := current
			for curr != nil {
				s := curr.step
				s.X = s.Node.X
				s.Y = s.Node.Y
				curr = curr.parent
				// If it's the first node of the path it has
				// no parent so we have to check it
				if curr != nil {
					curr.step.Node.NextStep = &Step{
						Node:   s.Node,
						Facing: s.Facing,
					}
				}
				s.Node = nil
				p = append(p, s)
				if atScale {
					if curr != nil {
						dx := s.X - curr.step.Node.X
						dy := s.Y - curr.step.Node.Y
						for i := 1; i < abs(dx); i++ {
							if dx > 0 {
								s.X -= 1
							} else {
								s.X += 1
							}
							p = append(p, s)

							// We check if the current step is the one we passed
							// as source so we can return as it as it was at Scale
							if s.X == sx && s.Y == sy {
								goto REVERSE_PATH
							}
						}
						for i := 1; i < abs(dy); i++ {
							if dy > 0 {
								s.Y -= 1
							} else {
								s.Y += 1
							}
							p = append(p, s)

							// We check if the current step is the one we passed
							// as source so we can return as it as it was at Scale
							if s.X == sx && s.Y == sy {
								goto REVERSE_PATH
							}
						}
					}
				}
			}
		REVERSE_PATH:
			for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
				p[i], p[j] = p[j], p[i]
			}

			return p
		}

		for _, neighbor := range current.step.Node.NeighborSteps {
			// If we know we are not gonna go to it just don't even push it to the queue
			// If it has a tower or is not a unit zone then don't use the node
			if neighbor.Node.HasTower() {
				continue
			}

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
				// If no node is ranked it's basically doing a dijkstra
				// which now it's much faster and gives better results
				//neighborStep.rank = cost + neighbor.Node.MDistance(tn)
				neighborStep.rank = cost
				neighborStep.parent = current
				// We try to give priority to the ones
				// that are not on the same direction.
				// So we increase the rank of the step
				// that is on the same direction as it
				// was before.
				// This helps on prioritizing doing
				// diagonals
				if neighborStep.step.Facing == current.step.Facing {
					neighborStep.rank += 32
				}
				heap.Push(nq, neighborStep)
			}
		}
	}
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

// checkConsecutiveSteps will check in the queueItem tree if it has
// 'c' consecutive steps before returning true or false
func checkConsecutiveSteps(qi *queueItem, c int) bool {
	if qi.parent != nil && qi.step.Node.NextStep != nil && qi.parent.step.Node.NextStep != nil && qi.parent.step.Node.NextStep.Node == qi.step.Node {
		c--
		if c == 0 {
			return true
		} else {
			return checkConsecutiveSteps(qi.parent, c)
		}
	}
	return false
}
