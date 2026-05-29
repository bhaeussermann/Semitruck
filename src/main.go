package main

import (
	"github.com/bhaeussermann/semitruck/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
  ebiten.SetWindowTitle("Semitruck")
  ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
  ebiten.SetWindowSize(1024, 768)
  ebiten.SetWindowSizeLimits(400, 350, -1, -1)
  ebiten.SetCursorMode(ebiten.CursorModeHidden)
  game, error := game.NewGame()
  if error != nil {
    panic(error)
  }
  error = ebiten.RunGame(game)
  if error != nil {
    panic(error)
  }
}
