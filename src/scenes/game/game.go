package game

import (
	"image/color"
	"log"
	"math"
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
	wheelImage *ebiten.Image

	isDisplayingMenu bool
	initializeMenu sync.Once
	menuBackgroundScale float32
	menu *menu.Menu
	exitToScene scenes.GetNextScene

	truck truck
}

type truck struct {
	spriteWidth float64
	spriteLength float64
	width float64
	length float64
	wheelDistance float64
	frontX float64
	frontY float64
	direction float64
	wheelTurnDirection float64
	speed float64
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
		truck: truck {
			spriteWidth: truckSpriteWidth,
			spriteLength: truckSpriteLength,
			width: widthRatio * truckSpriteWidth,
			length: (rearRatio - frontRatio) * truckSpriteLength,
			wheelDistance: truckSpriteLength * (rearWheelLengthRatio - frontWheelLengthRatio),
			frontX: 350,
			frontY: 150,
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
	}
	
	g.moveTruck()

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

func (g *Game) moveTruck() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.truck.speed = math.Min(maximumForwardSpeed, g.truck.speed + acceleration)
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		if g.truck.speed >= acceleration {
			g.truck.speed *= breakFactor
		} else {
			g.truck.speed = math.Max(-maximumReverseSpeed, g.truck.speed - acceleration)
		}
	} else {
		g.truck.speed *= frictionFactor
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.truck.wheelTurnDirection = math.Max(-maxTurnAngle, g.truck.wheelTurnDirection - turnSpeed)
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.truck.wheelTurnDirection = math.Min(maxTurnAngle, g.truck.wheelTurnDirection + turnSpeed)
	} else if math.Abs(g.truck.wheelTurnDirection) < turnSpeed {
		g.truck.wheelTurnDirection = 0
	} else if g.truck.wheelTurnDirection > 0 {
		g.truck.wheelTurnDirection -= turnSpeed
	} else {
		g.truck.wheelTurnDirection += turnSpeed
	}

	if math.Abs(g.truck.speed) > epsilon {
		rearWheelX := g.truck.frontX - g.truck.length * rearWheelLengthRatio * math.Cos(g.truck.direction)
		rearWheelY := g.truck.frontY - g.truck.length * rearWheelLengthRatio * math.Sin(g.truck.direction)

		if math.Abs(g.truck.wheelTurnDirection) > epsilon  {
			effectiveTurnAngle := math.Asin(g.truck.speed*math.Sin(g.truck.wheelTurnDirection * 2) / math.Sqrt(g.truck.speed*g.truck.speed + g.truck.wheelDistance*g.truck.wheelDistance + 2*g.truck.speed*g.truck.length*math.Cos(g.truck.wheelTurnDirection * 2)))
			g.truck.direction += effectiveTurnAngle
		}

		rearWheelX += g.truck.speed * math.Cos(g.truck.direction)
		rearWheelY += g.truck.speed * math.Sin(g.truck.direction)
		g.truck.frontX = rearWheelX + g.truck.length * rearWheelLengthRatio * math.Cos(g.truck.direction)
		g.truck.frontY = rearWheelY + g.truck.length * rearWheelLengthRatio * math.Sin(g.truck.direction)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.isDisplayingMenu {
		g.menuBackgroundScale = MaxF32(menuBackgroundTargetScale, g.menuBackgroundScale - menuBackgroundFadeSpeed)
	} else {
		g.menuBackgroundScale = MinF32(1, g.menuBackgroundScale + menuBackgroundFadeSpeed)
	}
	colorScale := ebiten.ColorScale{}
	colorScale.Scale(g.menuBackgroundScale, g.menuBackgroundScale, g.menuBackgroundScale, 1)

	g.drawRoad(screen, colorScale)
	g.drawTruck(screen, colorScale)

	if g.isDisplayingMenu {
		g.menu.Draw(screen)
	}
}

func MaxF32(x float32, y float32) float32 {
	if x >= y { return x } else { return y }
}

func MinF32(x float32, y float32) float32 {
	if x <= y { return x } else { return y }
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

var acceleration = 0.03
var frictionFactor = 0.95
var breakFactor = 0.9
var maximumForwardSpeed = 3.0
var maximumReverseSpeed = 1.5
var maxTurnAngle = math.Pi / 4
var turnSpeed = maxTurnAngle / 10

var widthRatio = 0.775
var frontRatio = 0.01
var rearRatio = 0.98
var frontWheelLengthRatio = 0.15
var rearWheelLengthRatio = 0.85
var wheelLength = 30
var wheelWidth = 12

var epsilon = 0.00001
