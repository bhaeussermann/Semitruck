package game

import (
	_ "image/jpeg"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	screenSize coordinates
	pointerFront coordinatesf
	pointerPivot coordinatesf
	pointerWidth float64
	pointerLength float64
	pointerImage *ebiten.Image
	backgroundImage *ebiten.Image
}

func NewGame() *Game {
	backgroundImage, _, error := ebitenutil.NewImageFromFile("../images/menu-background.jpg")
	if error != nil {
		log.Fatal(error)
	}

	pointerImage, _, error := ebitenutil.NewImageFromFile("../images/pointer.png")
	if error != nil {
		log.Fatal(error)
	}

	return &Game{
		pointerFront: coordinatesf{x: 200, y: 200},
		pointerPivot: nilCoordinates,
		pointerWidth: float64(pointerImage.Bounds().Dy()),
		pointerLength: float64(pointerImage.Bounds().Dx()),
		pointerImage: pointerImage,
		backgroundImage: backgroundImage,
	}
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	g.screenSize = coordinates{x: outsideWidth, y: outsideHeight}
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	cursorX, cursorY := ebiten.CursorPosition()
	if (cursorX >= 0) && (cursorY >= 0) && (cursorX < g.screenSize.x) && (cursorY < g.screenSize.y) {
		g.pointerFront = coordinatesf{x: float64(cursorX), y: float64(cursorY)}

		pointerAngle := math.Atan2(g.pointerFront.y - g.pointerPivot.y, g.pointerFront.x - g.pointerPivot.x)
		pointerPivotLocation := g.pointerLength * pointerPivotLengthRatio
		g.pointerPivot = coordinatesf{
			x: g.pointerFront.x - pointerPivotLocation * math.Cos(pointerAngle),
			y: g.pointerFront.y - pointerPivotLocation * math.Sin(pointerAngle),
		}
	} else {
		g.pointerFront = nilCoordinates
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)

	if g.pointerFront != nilCoordinates {
		g.drawArrow(screen)
	}
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	screenWidth := float64(screen.Bounds().Dx())
	screenHeight := float64(screen.Bounds().Dy())
	imageWidth := float64(g.backgroundImage.Bounds().Dx())
	imageHeight := float64(g.backgroundImage.Bounds().Dy())

	horizontalScale := screenWidth / imageWidth
	verticalScale := screenHeight / imageHeight
	scale := math.Max(horizontalScale, verticalScale)

	imageDrawGeom := ebiten.GeoM{}
	imageDrawGeom.Translate(-imageWidth / 2, -imageHeight / 2)
	imageDrawGeom.Scale(scale, scale)
	imageDrawGeom.Translate(screenWidth / 2, screenHeight / 2)
	screen.DrawImage(g.backgroundImage, &ebiten.DrawImageOptions{GeoM: imageDrawGeom})
}

func (g *Game) drawArrow(screen *ebiten.Image) {
	angle := math.Atan2(g.pointerFront.y - g.pointerPivot.y, g.pointerFront.x - g.pointerPivot.x)

	drawGeom := ebiten.GeoM{}
	drawGeom.Translate(-g.pointerLength, -g.pointerWidth/2)
	drawGeom.Rotate(angle)
	drawGeom.Translate(g.pointerFront.x, g.pointerFront.y)
	screen.DrawImage(g.pointerImage, &ebiten.DrawImageOptions{GeoM: drawGeom})
}

type coordinates struct {
	x int
	y int
}

type coordinatesf struct {
	x float64
	y float64
}

var nilCoordinates = coordinatesf{x: -1, y: -1}

var pointerPivotLengthRatio = float64(0.5)
