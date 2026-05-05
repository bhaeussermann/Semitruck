package game

import (
	"log"
	"sync"

	"github.com/bhaeussermann/semitruck/components"
	"github.com/bhaeussermann/semitruck/components/menu"
	"github.com/bhaeussermann/semitruck/scenes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	screenSize components.Coordinates
	roadImage *ebiten.Image
	truckImage *ebiten.Image

	isDisplayingMenu bool
	initializeMenu sync.Once
	menu *menu.Menu
	exitToScene scenes.GetNextScene

	truck truck
}

type truck struct {
	x float64
	y float64
}

func NewGame(exitToScene scenes.GetNextScene) (scenes.Scene, error) {
	roadImage, _, error := ebitenutil.NewImageFromFile("../images/road.png")
	if error != nil {
		return nil, error
	}

	truckImage, _, error := ebitenutil.NewImageFromFile("../images/truck.png")
	if error != nil {
		return nil, error
	}

	return &Game{
		roadImage: roadImage,
		truckImage: truckImage,
		exitToScene: exitToScene,
		truck: truck {
			x: 100,
			y: 50,
		},
	}, nil
}

func (g *Game) SetScreenSize(width int, height int) {
	g.screenSize = components.Coordinates{X: width, Y: height}
	if g.isDisplayingMenu {
		g.menu.SetScreenSize(width, height)
	}
}

func (g *Game) Update() scenes.SceneChange {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.isDisplayingMenu = !g.isDisplayingMenu
		if g.isDisplayingMenu {
			g.initializeMenu.Do(g.createMenu)
		}
	}

	if g.isDisplayingMenu {
		menuItemSelectionIndex := g.menu.Update()
		switch menuItemSelectionIndex {
		case 0:
			g.isDisplayingMenu = false
		case 1:
			return scenes.SceneChange{GetNextScene: g.exitToScene}
		default:
			return scenes.SceneChange{}
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.truck.x += 0.2
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.truck.x -= 0.2
	}

	return scenes.SceneChange{}
}

func (g *Game) createMenu() {
	menuItemTexts := []string{
		"Continue",
		"Exit to menu",
	}
	var error error
	g.menu, error = menu.NewMenu(menuItemTexts)
	if error != nil {
		log.Fatal(error)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	colorScale := ebiten.ColorScale{}
	if g.isDisplayingMenu {
		colorScale.Scale(0.5, 0.5, 0.5, 1)
	}

	g.drawRoad(screen, colorScale)
	g.drawTruck(screen, colorScale)

	if g.isDisplayingMenu {
		g.menu.Draw(screen)
	}
}

func (g *Game) drawRoad(screen *ebiten.Image, colorScale ebiten.ColorScale) {
	for x := 0; x < g.screenSize.X; x += g.roadImage.Bounds().Dx() {
		for y := 0; y < g.screenSize.Y; y += g.roadImage.Bounds().Dy() {
			geom := ebiten.GeoM{}
			geom.Translate(float64(x), float64(y))
			screen.DrawImage(g.roadImage, &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geom})
		}
	}
}

func (g *Game) drawTruck(screen *ebiten.Image, colorScale ebiten.ColorScale) {
	geom := ebiten.GeoM{}
	geom.Translate(g.truck.x, g.truck.y)
	screen.DrawImage(g.truckImage, &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geom})
}
