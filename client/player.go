package main

type Player struct {
	Line  *Line
	Lives int
}

func NewPlayer(position int) *Player {
	p := &Player{
		Lives: 20,
	}
	p.Line = NewLine(p)
	return p
}

func (p *Player) Update() error {
	p.Line.Update()
	return nil
}

//func (p *Player) Draw(screen *ebiten.Image, c *Camera) {
//p.Line.Draw(screen, c)
//}
