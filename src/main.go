package main

import (
	"log"

	"github.com/bhaeussermann/semitruck/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
  ebiten.SetWindowTitle("Semitruck")
  ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
  ebiten.SetWindowSize(800, 600)
  ebiten.SetCursorMode(ebiten.CursorModeHidden)
  error := ebiten.RunGame(game.NewGame())
  if error != nil {
    log.Fatal(error)
  }
}
