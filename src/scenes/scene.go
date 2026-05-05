package scenes

import "github.com/hajimehoshi/ebiten/v2"

type Scene interface {
	SetScreenSize(width int, height int)
	Update() SceneChange
	Draw(screen *ebiten.Image)
}

type SceneChange struct {
	Terminate bool
	GetNextScene GetNextScene
}

type GetNextScene func() (Scene, error)
