package game

import (
	"log"

	"github.com/bhaeussermann/semitruck/scenes"
	gamemenu "github.com/bhaeussermann/semitruck/scenes/game-menu"
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	currentScene scenes.Scene
}

func NewGame() (*Game, error) {
	scene, error := gamemenu.NewGameMenu()
	if error != nil {
		return nil, error
	}
	
	return &Game {
		currentScene: scene,
	}, nil
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	g.currentScene.SetScreenSize(outsideWidth, outsideHeight)
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	sceneChange := g.currentScene.Update()
	if sceneChange.Terminate {
		return ebiten.Termination
	}
	if sceneChange.GetNextScene != nil {
		nextScene, error := sceneChange.GetNextScene()
		if error != nil {
			log.Fatal(error)
		} else {
			g.currentScene = nextScene
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.currentScene.Draw(screen)
}
