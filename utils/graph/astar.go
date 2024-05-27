package graph

import (
	"container/heap"
	"math"

	"github.com/xescugc/maze-wars/unit/environment"
	"github.com/xescugc/maze-wars/utils"
)

const (
	basicTPS float64 = 60
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
// to Target(tx,ty) with the Movement Speed(ms) and starting on the Direction(d)
// with W,H equal to the Scale in the designed Environment(env).
// If atScale is true it'll return the 1:1 result, if not it'll return
// the 1:Scale result
func (g *Graph) AStar(sx, sy, ms float64, d utils.Direction, tx, ty int, env environment.Environment, atScale bool) []Step {
	nm := stepMap{}
	nq := &queue{}
	heap.Init(nq)
	var (
		sn, tn *Node
	)
	if atScale {
		sn = g.GetNodeOf(int(sx), int(sy))
		tn = g.GetNodeOf(tx, ty)
	} else {
		sn = g.GetNode(int(sx), int(sy))
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

		// We want to enter the return loop, where we calculate the end path to return if:
		// * The current node is the end node
		// * There are 4 consecutive already known steps, then we use the cache on Node.NextStep
		// * It's aerial, which means it just goes straight down, noting to calculate
		if current.step.Node.ID == tn.ID || checkConsecutiveSteps(current, 4) || env == environment.Aerial {
			if env == environment.Aerial {
				// If it's an Aerial environment it has to go straight down until
				// the next node is Death Zone
				current = &queueItem{
					step: Step{
						Node:   current.step.Node.BottomNeighbor,
						Facing: d,
					},
					parent: current,
				}
				for !current.step.Node.IsDeathZone {
					current = &queueItem{
						step: Step{
							Node:   current.step.Node.BottomNeighbor,
							Facing: d,
						},
						parent: current,
					}
				}

			} else {
				// If it has a NextStep then it builds it up from the
				// cache that is NextStep by following it up until the end
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
			}
			// Found a path to the goal.
			// And it's on reverse order
			p := []Step{}
			curr := current
			for curr != nil {
				s := curr.step
				s.X = float64(s.Node.X)
				s.Y = float64(s.Node.Y)
				curr = curr.parent
				// If it's the first node of the path it has
				// no parent so we have to check it
				if curr != nil && env != environment.Aerial {
					// TODO: if it's Aerial don't do this
					curr.step.Node.NextStep = &Step{
						Node:   s.Node,
						Facing: s.Facing,
					}
				}
				s.Node = nil
				p = append(p, s)
				if atScale {
					if curr != nil {
						dx := s.X - float64(curr.step.Node.X)
						dy := s.Y - float64(curr.step.Node.Y)

						// We calculate the number of movements needed to reach
						// with the MS defined and on the basicTPS
						msdx := (absF(dx) * basicTPS) / ms
						msdy := (absF(dy) * basicTPS) / ms

						// We calculate the actual distance it has to move to reach
						// the position with the MS
						distx := absF(dx) / msdx
						disty := absF(dy) / msdy
						// As diagonal moves do not exist I just need
						// to move the difference between nodes in
						// X and Y, which is the DX and DY
						// Moving in X position
						for i := 1; i < int(absF(math.Round(msdx))); i++ {
							if dx > 0 {
								s.X -= distx
							} else {
								s.X += distx
							}
							p = append(p, s)

							// We check if the current step is the one we passed
							// as source so we can return as it as it was at Scale
							if math.Round(s.X) == sx && math.Round(s.Y) == sy {
								goto REVERSE_PATH
							}
						}
						// Moving in Y position
						for i := 1; i < int(absF(math.Round(msdy))); i++ {
							if dy > 0 {
								s.Y -= disty
							} else {
								s.Y += disty
							}
							p = append(p, s)

							// We check if the current step is the one we passed
							// as source so we can return as it as it was at Scale
							if math.Round(s.X) == sx && math.Round(s.Y) == sy {
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
