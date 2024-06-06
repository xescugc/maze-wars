package bot

import (
	"context"
	"math/rand"
	"sort"
	"time"

	bht "github.com/joeycumines/go-behaviortree"
	"github.com/xescugc/go-flux"
	"github.com/xescugc/maze-wars/action"
	"github.com/xescugc/maze-wars/store"
	"github.com/xescugc/maze-wars/tower"
	"github.com/xescugc/maze-wars/unit"
)

type Bot struct {
	Ticker bht.Ticker

	dispatcher *flux.Dispatcher
	store      *store.Store
	playerID   string

	context     context.Context
	ctxCancelFn context.CancelFunc

	towerIDToUpdate   string
	towerTypeToUpdate string
}

func New(ctx context.Context, d *flux.Dispatcher, s *store.Store, pid string) *Bot {
	b := &Bot{
		dispatcher: d,
		store:      s,
		playerID:   pid,
	}
	b.context, b.ctxCancelFn = context.WithCancel(ctx)

	return b
}

func (b *Bot) Start() {
	go func() {
		ticker := bht.NewTicker(b.context, time.Second/4, b.Node())
		<-ticker.Done()
	}()
}

func (b *Bot) Stop() {
	b.ctxCancelFn()
}

func (b *Bot) Node() bht.Node {
	return bht.New(
		bht.Shuffle(bht.Selector, nil),
		//Units
		bht.New(
			bht.Shuffle(bht.Selector, nil),
			// Update
			bht.New(
				bht.Selector,
				b.updateUnits()...,
			),
			// Summon
			bht.New(
				bht.Shuffle(bht.Selector, nil),
				b.summonUnits()...,
			),
		),
		// Towers
		bht.New(
			bht.Shuffle(bht.Selector, nil),
			// Update
			bht.New(
				bht.Sequence,
				bht.New(b.findTowerToUpdate()),
				bht.New(b.updateTower()),
			),
			// Place
			bht.New(
				bht.Shuffle(bht.Selector, nil),
				b.placeTowers()...,
			),
		),
	)
}

func (b *Bot) updateUnits() []bht.Node {
	res := make([]bht.Node, 0, 0)
	units := make([]*unit.Unit, 0, len(unit.Units))
	for _, u := range unit.Units {
		units = append(units, u)
	}
	sort.Slice(units, func(i, j int) bool { return units[i].Gold < units[j].Gold })
	for _, u := range units {
		res = append(res, bht.New(
			bht.Sequence,
			bht.New(b.canUpdateUnit(u.Type.String())),
			bht.New(b.updateUnit(u.Type.String())),
		))
	}
	return res
}

func (b *Bot) canUpdateUnit(ut string) func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		cp := b.store.Lines.FindPlayerByID(b.playerID)
		if cp.CanUpdateUnit(ut) {
			return bht.Success, nil
		} else {
			return bht.Failure, nil
		}
	}
}

func (b *Bot) updateUnit(ut string) func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		b.dispatcher.Dispatch(action.NewUpdateUnit(b.playerID, ut))
		return bht.Success, nil
	}
}

func (b *Bot) summonUnits() []bht.Node {
	res := make([]bht.Node, 0, 0)
	for _, u := range unit.Units {
		res = append(res, bht.New(
			bht.Sequence,
			bht.New(b.canSummonUnit(u.Type.String())),
			bht.New(b.summonUnit(u.Type.String())),
		))
	}
	return res
}

func (b *Bot) canSummonUnit(ut string) func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		cp := b.store.Lines.FindPlayerByID(b.playerID)
		if cp.CanSummonUnit(ut) {
			return bht.Success, nil
		} else {
			return bht.Failure, nil
		}
	}
}

func (b *Bot) summonUnit(ut string) func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		cp := b.store.Lines.FindPlayerByID(b.playerID)
		nlid := b.store.Map.GetNextLineID(cp.LineID)

		b.dispatcher.Dispatch(action.NewSummonUnit(ut, cp.ID, cp.LineID, nlid))

		return bht.Success, nil
	}
}

func (b *Bot) findTowerToUpdate() func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		cp := b.store.Lines.FindPlayerByID(b.playerID)
		for _, t := range b.store.Lines.FindLineByID(cp.LineID).Towers {
			tus := tower.Towers[t.Type].Updates
			if len(tus) == 0 {
				continue
			}
			i := rand.Intn(len(tus))
			tu := tower.Towers[tus[i].String()]
			if cp.Gold-tu.Gold > 0 {
				b.towerIDToUpdate = t.ID
				b.towerTypeToUpdate = tu.Type.String()
				return bht.Success, nil
			}
		}
		return bht.Failure, nil
	}
}

func (b *Bot) updateTower() func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		b.dispatcher.Dispatch(action.NewUpdateTower(b.playerID, b.towerIDToUpdate, b.towerTypeToUpdate))
		b.towerIDToUpdate = ""
		b.towerTypeToUpdate = ""
		return bht.Success, nil
	}
}

func (b *Bot) placeTowers() []bht.Node {
	res := make([]bht.Node, 0, 0)
	for _, t := range tower.FirstTowers {
		res = append(res, bht.New(
			bht.Sequence,
			bht.New(b.canPlaceTower(t.Type.String())),
			bht.New(b.placeTower(t.Type.String())),
		))
	}
	return res
}

func (b *Bot) canPlaceTower(tt string) func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		cp := b.store.Lines.FindPlayerByID(b.playerID)
		if cp.CanPlaceTower(tt) {
			return bht.Success, nil
		} else {
			return bht.Failure, nil
		}
	}
}

func (b *Bot) placeTower(tt string) func(children []bht.Node) (bht.Status, error) {
	return func(children []bht.Node) (bht.Status, error) {
		cp := b.store.Lines.FindPlayerByID(b.playerID)
		cl := b.store.Lines.FindLineByID(cp.LineID)

		x, y := b.store.Map.GetHomeCoordinates(cp.LineID)
		x += 16                                                          // Move one tile away so we are inside and not in the border
		y += 16 + (cl.Graph.SpawnZoneH * cl.Graph.Scale) + 16 + (5 * 16) // Top border + H zone + Go to the next one + (10 tiles of the top)

		tl := len(cl.Towers) / 7
		ts := len(cl.Towers) % 7

		// The maximum number of lines that can be fitted on the line
		if tl == 23 {
			return bht.Failure, nil
		}
		// For each line of tower we add the difference
		y += tl * (32 + 16)

		// If it's odd we start from the left
		// If it's even we start from the right
		// We do a +1 so 0 is 1 and then it's the
		// odd-even play
		if (tl+1)%2 == 0 {
			// We take it to the end
			x += 32 * 7
			x -= 32 * ts
		} else {
			x += 32 * ts
		}

		// There are max of 7 towers in one line
		// we'll leave a 10 tile margin from the top
		b.dispatcher.Dispatch(action.NewPlaceTower(tt, cp.ID, x, y))

		return bht.Success, nil
	}
}
