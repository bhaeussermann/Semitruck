package game

import (
	"image/color"
	"log"
	"math"
	"slices"
	"sync"

	"github.com/bhaeussermann/semitruck/components"
	"github.com/bhaeussermann/semitruck/components/menu"
	"github.com/bhaeussermann/semitruck/scenes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct {
	screenSize components.Coordinates
	roadImage *ebiten.Image
	truckImage *ebiten.Image
	wheelImage *ebiten.Image

	isDisplayingMenu bool
	initializeMenu sync.Once
	menuBackgroundScale float32
	menu *menu.Menu
	exitToScene scenes.GetNextScene

	borders []*border
	truck *truck
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
	truckSpriteWidth := float64(truckImage.Bounds().Dy())
	truckSpriteLength := float64(truckImage.Bounds().Dx())

	wheelImage := ebiten.NewImage(wheelLength, wheelWidth)
	wheelImage.Fill(color.Black)

	return &Game{
		roadImage: roadImage,
		truckImage: truckImage,
		wheelImage: wheelImage,
		exitToScene: exitToScene,
		menuBackgroundScale: 1,
		borders: []*border{
			createBorder([]point{{100, 450}, {400, 300}, {550, 280}, {550, 120}}, 20),
			createBorder([]point{{350, 620}, {750, 620}}, 20),
		},
		truck: &truck{
			spriteWidth: truckSpriteWidth,
			spriteLength: truckSpriteLength,
			width: widthRatio * truckSpriteWidth,
			length: (rearRatio - frontRatio) * truckSpriteLength,
			wheelDistance: truckSpriteLength * (rearWheelLengthRatio - frontWheelLengthRatio),
			frontX: 350,
			frontY: 150,
		},
		screenSize: components.Coordinates{},
		isDisplayingMenu: false,
		initializeMenu: sync.Once{},
		menu: &menu.Menu{},
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
	}

	g.truck.updateMovement()

	g.computeCollisions()

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

func (g *Game) computeCollisions() {
	truck := g.truck

	frontLeftX := truck.frontX + math.Sin(truck.direction) * truck.width / 2
	frontLeftY := truck.frontY - math.Cos(truck.direction) * truck.width / 2
	rearLeftX := frontLeftX - math.Cos(truck.direction) * truck.length
	rearLeftY := frontLeftY - math.Sin(truck.direction) * truck.length
	frontRightX := truck.frontX - math.Sin(truck.direction) * truck.width / 2
	frontRightY := truck.frontY + math.Cos(truck.direction) * truck.width / 2
	rearRightX := frontRightX - math.Cos(truck.direction) * truck.length
	rearRightY := frontRightY - math.Sin(truck.direction) * truck.length

	truckEdges := []*edgeLine{
		createEdgeLine(frontLeftX, frontLeftY, frontRightX, frontRightY),
		createEdgeLine(rearLeftX, rearLeftY, rearRightX, rearRightY),
		createEdgeLine(frontLeftX, frontLeftY, rearLeftX, rearLeftY),
		createEdgeLine(frontRightX, frontRightY, rearRightX, rearRightY),
	}

	for _, courseEdge := range g.getAllEdges() {
		for _, truckEdge := range truckEdges {
			if truckEdge.intersectsEdge(courseEdge) {
				truck.bump(truckEdge, courseEdge)
				return
			}
		}
	}
}

func (g *Game) getAllEdges() []*edge {
	edges := []*edge{}
	for _, border := range g.borders {
		edges = slices.Concat(edges, border.edges)
	}
	return edges
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.isDisplayingMenu {
		g.menuBackgroundScale = MaxF32(menuBackgroundTargetScale, g.menuBackgroundScale-menuBackgroundFadeSpeed)
	} else {
		g.menuBackgroundScale = MinF32(1, g.menuBackgroundScale+menuBackgroundFadeSpeed)
	}
	colorScale := ebiten.ColorScale{}
	colorScale.Scale(g.menuBackgroundScale, g.menuBackgroundScale, g.menuBackgroundScale, 1)

	g.drawRoad(screen, colorScale)
	for _, border := range g.borders {
		drawBorder(screen, border, colorScale)
	}
	g.drawTruck(screen, colorScale)

	if g.isDisplayingMenu {
		g.menu.Draw(screen)
	}
}

func MaxF32(x float32, y float32) float32 {
	if x >= y {
		return x
	} else {
		return y
	}
}

func MinF32(x float32, y float32) float32 {
	if x <= y {
		return x
	} else {
		return y
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

func drawBorder(screen *ebiten.Image, border *border, colorScale ebiten.ColorScale) {
	path := vector.Path{}
	startLine := border.edges[0].centerLine
	path.MoveTo(float32(startLine.x1), float32(startLine.y1))
	for _, borderEdge := range border.edges {
		path.LineTo(float32(borderEdge.centerLine.x2), float32(borderEdge.centerLine.y2))
	}
	vector.StrokePath(
		screen,
		&path,
		&vector.StrokeOptions{Width: float32(border.width)},
		&vector.DrawPathOptions{AntiAlias: true, ColorScale: colorScale})
}

func (g *Game) drawTruck(screen *ebiten.Image, colorScale ebiten.ColorScale) {
	g.drawTruckWheel(screen, -g.truck.width / 2 + float64(wheelWidth) / 2)
	g.drawTruckWheel(screen, g.truck.width / 2 - float64(wheelWidth) / 2)

	geom := ebiten.GeoM{}
	geom.Translate(-g.truck.spriteLength * (1 - frontRatio), -g.truck.spriteWidth / 2)
	geom.Rotate(g.truck.direction)
	geom.Translate(g.truck.frontX, g.truck.frontY)
	screen.DrawImage(g.truckImage, &ebiten.DrawImageOptions{ColorScale: colorScale, GeoM: geom})
}

func (g *Game) drawTruckWheel(screen *ebiten.Image, offset float64) {
	geom := ebiten.GeoM{}
	geom.Translate(-float64(wheelLength) / 2, -float64(wheelWidth) / 2)
	geom.Rotate(g.truck.wheelTurnDirection)
	geom.Translate(-g.truck.length * frontWheelLengthRatio, offset)
	geom.Rotate(g.truck.direction)
	geom.Translate(g.truck.frontX, g.truck.frontY)
	screen.DrawImage(g.wheelImage, &ebiten.DrawImageOptions{GeoM: geom})
}

var menuBackgroundTargetScale float32 = 0.5
var menuBackgroundFadeSpeed float32 = 0.05

var widthRatio = 0.775
var frontRatio = 0.01
var rearRatio = 0.98
var frontWheelLengthRatio = 0.15
var rearWheelLengthRatio = 0.85
var wheelLength = 30
var wheelWidth = 12
