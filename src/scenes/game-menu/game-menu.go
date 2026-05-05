package gamemenu

import (
	"bytes"
	_ "image/jpeg"
	"math"

	"github.com/bhaeussermann/semitruck/components/menu"
	"github.com/bhaeussermann/semitruck/scenes"
	"github.com/bhaeussermann/semitruck/scenes/game"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
)

type GameMenu struct {
	menu *menu.Menu
	backgroundImage *ebiten.Image
	textFaceSource *text.GoTextFaceSource
}

func NewGameMenu() (scenes.Scene, error) {
	backgroundImage, _, error := ebitenutil.NewImageFromFile("../images/menu-background.jpg")
	if error != nil {
		return nil, error
	}

	textFaceSource, error := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if error != nil {
		return nil, error
	}

	menu, error := menu.NewMenu(menuItemTexts)
	if error != nil {
		return nil, error
	}

	return &GameMenu{
		menu: menu,
		backgroundImage: backgroundImage,
		textFaceSource: textFaceSource,
	}, nil
}

func (m *GameMenu) SetScreenSize(width int, height int) {
	m.menu.SetScreenSize(width, height)
}

func (m *GameMenu) Update() scenes.SceneChange {
	menuItemSelectionIndex := m.menu.Update()
	switch menuItemSelectionIndex {
	case 0: return scenes.SceneChange{GetNextScene: func() (scenes.Scene, error) { return game.NewGame(NewGameMenu) }}
	case 1: return scenes.SceneChange{Terminate: true}
	default: return scenes.SceneChange{}
	}
}

func (m *GameMenu) Draw(screen *ebiten.Image) {
	m.drawBackground(screen)
	m.drawTitle(screen)
	m.menu.Draw(screen)
}

func (m *GameMenu) drawBackground(screen *ebiten.Image) {
	screenWidth := float64(screen.Bounds().Dx())
	screenHeight := float64(screen.Bounds().Dy())
	imageWidth := float64(m.backgroundImage.Bounds().Dx())
	imageHeight := float64(m.backgroundImage.Bounds().Dy())

	horizontalScale := screenWidth / imageWidth
	verticalScale := screenHeight / imageHeight
	scale := math.Max(horizontalScale, verticalScale)

	imageDrawGeom := ebiten.GeoM{}
	imageDrawGeom.Translate(-imageWidth / 2, -imageHeight / 2)
	imageDrawGeom.Scale(scale, scale)
	imageDrawGeom.Translate(screenWidth / 2, screenHeight / 2)
	screen.DrawImage(m.backgroundImage, &ebiten.DrawImageOptions{GeoM: imageDrawGeom})
}

func (m *GameMenu) drawTitle(screen *ebiten.Image) {
	textFace := text.GoTextFace{
		Source: m.textFaceSource,
		Size: titleTextSize,
	}
	drawGeom := ebiten.GeoM{}
	drawGeom.Translate(float64(screen.Bounds().Dx()) / 2, 20)
	drawOptions := &text.DrawOptions {
		DrawImageOptions: ebiten.DrawImageOptions{
			GeoM: drawGeom,
		},
		LayoutOptions: text.LayoutOptions{
			PrimaryAlign: text.AlignCenter,
		},
	}
	text.Draw(screen, "Semitruck", &textFace, drawOptions)
}

var menuItemTexts = []string{
	"Start game",
	"Exit",
}

var titleTextSize = float64(42)
