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
		borders: []*border{
			createBorder([]point{{100, 450}, {400, 300}, {550, 280}, {550, 120}}, 20),
			createBorder([]point{{250, 520}, {650, 520}}, 20),
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

	g.truck.move()

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

func (t *truck) move() {
	t.updateSpeed()
	t.turn()
}

func (t *truck) updateSpeed() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		t.speed = math.Min(maximumForwardSpeed, t.speed + acceleration)
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		if t.speed >= acceleration {
			t.speed *= breakFactor
		} else {
			t.speed = math.Max(-maximumReverseSpeed, t.speed - acceleration)
		}
	} else {
		t.speed *= frictionFactor
	}
}

func (t *truck) turn() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		t.wheelTurnDirection = math.Max(-maxTurnAngle, t.wheelTurnDirection - turnSpeed)
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		t.wheelTurnDirection = math.Min(maxTurnAngle, t.wheelTurnDirection + turnSpeed)
	} else if math.Abs(t.wheelTurnDirection) < turnSpeed {
		t.wheelTurnDirection = 0
	} else if t.wheelTurnDirection > 0 {
		t.wheelTurnDirection -= turnSpeed
	} else {
		t.wheelTurnDirection += turnSpeed
	}

	if math.Abs(t.speed) > epsilon {
		rearWheelX := t.frontX - t.length*rearWheelLengthRatio*math.Cos(t.direction)
		rearWheelY := t.frontY - t.length*rearWheelLengthRatio*math.Sin(t.direction)

		if math.Abs(t.wheelTurnDirection) > epsilon {
			effectiveTurnAngle := math.Asin(t.speed * math.Sin(t.wheelTurnDirection*2) / math.Sqrt(t.speed * t.speed + t.wheelDistance * t.wheelDistance + 2 * t.speed * t.length * math.Cos(t.wheelTurnDirection * 2)))
			t.direction += effectiveTurnAngle
		}

		rearWheelX += t.speed * math.Cos(t.direction)
		rearWheelY += t.speed * math.Sin(t.direction)
		t.frontX = rearWheelX + t.length*rearWheelLengthRatio * math.Cos(t.direction)
		t.frontY = rearWheelY + t.length*rearWheelLengthRatio * math.Sin(t.direction)
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

	truckEdges := []*edgeLine {
		createEdgeLine(frontLeftX, frontLeftY, frontRightX, frontRightY),
		createEdgeLine(frontLeftX, frontLeftY, rearLeftX, rearLeftY),
		createEdgeLine(frontRightX, frontRightY, rearRightX, rearRightY),
		createEdgeLine(rearLeftX, rearLeftY, rearRightX, rearRightY),
	}

	gameEdges := g.getAllEdges()
	for _, edge := range truckEdges {
		if slices.ContainsFunc(gameEdges, edge.intersectsEdge) {
			truck.speed = 0
			return
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
		g.menuBackgroundScale = MaxF32(menuBackgroundTargetScale, g.menuBackgroundScale - menuBackgroundFadeSpeed)
	} else {
		g.menuBackgroundScale = MinF32(1, g.menuBackgroundScale + menuBackgroundFadeSpeed)
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
		&vector.StrokeOptions{ Width: float32(border.width) },
		&vector.DrawPathOptions{ AntiAlias: true, ColorScale: colorScale })
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
