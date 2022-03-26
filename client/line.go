package main

type Line struct {
	Player  *Player
	Enemies []*Enemy
	Towers  []*Tower
}

func NewLine(p *Player) *Line {
	return &Line{
		Player: p,
	}
}

func (l *Line) Update() error {
	for _, e := range l.Enemies {
		e.Update()
	}
	return nil
}

//func (l *Line) Draw(screen *ebiten.Image, c *Camera) {
//for _, e := range l.Enemies {
//e.Draw(screen, c)
//}
//}
