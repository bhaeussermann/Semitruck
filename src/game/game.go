package game

import "github.com/hajimehoshi/ebiten/v2"

type Game struct{}

func NewGame() *Game {
  return &Game{}
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
  return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
  return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}
