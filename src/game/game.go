package game

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"log"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
)

type Game struct {
	screenSize coordinates
	pointerFront coordinatesf
	pointerPivot coordinatesf
	selectedMenuItemIndex int
	menuItems []*menuItem

	pointerWidth float64
	pointerLength float64
	pointerImage *ebiten.Image
	backgroundImage *ebiten.Image
	textFaceSource *text.GoTextFaceSource
}

type menuItem struct {
	text string
	animationEnd time.Time
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

	textFaceSource, error := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if error != nil {
		log.Fatal(error)
	}

	menuItems := make([]*menuItem, len(menuItemTexts))
	for menuItemIndex, menuItemText := range menuItemTexts {
		menuItems[menuItemIndex] = &menuItem{ text: menuItemText, }
	}

	return &Game{
		pointerFront: coordinatesf{200, 200},
		pointerPivot: nilCoordinates,
		selectedMenuItemIndex: -1,
		menuItems: menuItems,

		pointerWidth: float64(pointerImage.Bounds().Dy()),
		pointerLength: float64(pointerImage.Bounds().Dx()),
		pointerImage: pointerImage,
		backgroundImage: backgroundImage,
		textFaceSource: textFaceSource,
	}
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	g.screenSize = coordinates{outsideWidth, outsideHeight}
	return outsideWidth, outsideHeight
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		if g.selectedMenuItemIndex == len(g.menuItems) - 1 {
			return ebiten.Termination
		}
	}

	cursorX, cursorY := ebiten.CursorPosition()
	if (cursorX >= 0) && (cursorY >= 0) && (cursorX < g.screenSize.x) && (cursorY < g.screenSize.y) {
		g.updateCursorPointer(cursorX, cursorY)
		g.updateMenuItemSelection(cursorX, cursorY)
	} else {
		g.pointerFront = nilCoordinates
	}

	return nil
}

func (g *Game) updateCursorPointer(cursorX int, cursorY int) {
	g.pointerFront = coordinatesf{float64(cursorX), float64(cursorY)}

	pointerAngle := math.Atan2(g.pointerFront.y - g.pointerPivot.y, g.pointerFront.x - g.pointerPivot.x)
	pointerPivotLocation := g.pointerLength * pointerPivotLengthRatio
	g.pointerPivot = coordinatesf{
		g.pointerFront.x - pointerPivotLocation * math.Cos(pointerAngle),
		g.pointerFront.y - pointerPivotLocation * math.Sin(pointerAngle),
	}
}

func (g *Game) updateMenuItemSelection(cursorX int, cursorY int) {
	nextSelectedMenuItemIndex := -1
	menuItemsBounds := g.getMenuItemBounds(g.getMenuItemFonts())
	for menuItemIndex, menuItemBounds := range menuItemsBounds {
		if (cursorX >= menuItemBounds.Min.X) && (cursorX <= menuItemBounds.Max.X) && (cursorY >= menuItemBounds.Min.Y) && (cursorY <= menuItemBounds.Max.Y) {
			nextSelectedMenuItemIndex = menuItemIndex
			break
		}
	}
	if nextSelectedMenuItemIndex != g.selectedMenuItemIndex {
		if g.selectedMenuItemIndex != -1 {
			g.updateMenuItemAnimation(g.selectedMenuItemIndex)
		}
		g.selectedMenuItemIndex = nextSelectedMenuItemIndex
		if g.selectedMenuItemIndex != -1 {
			g.updateMenuItemAnimation(g.selectedMenuItemIndex)
		}
	}
}

func (g *Game) updateMenuItemAnimation(menuItemIndex int) {
	menuItem := g.menuItems[menuItemIndex]
	animationEndFromNow := time.Now().Add(menuItemTextAnimationDuration)
	isAnimating := !menuItem.animationEnd.IsZero() && menuItem.animationEnd.After(time.Now())
	if isAnimating {
		elapsedAnimationDuration := animationEndFromNow.Sub(menuItem.animationEnd)
		animationProgress := float64(elapsedAnimationDuration) / float64(menuItemTextAnimationDuration)
		newAnimationProgress := 1 - math.Sqrt(animationProgress * (2 - animationProgress))
		menuItem.animationEnd = time.Now().Add(time.Duration(int64((1 - newAnimationProgress) * float64(menuItemTextAnimationDuration))))
	} else {
		menuItem.animationEnd = animationEndFromNow
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.drawTitle(screen)
	g.drawMenu(screen)

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

func (g *Game) drawTitle(screen *ebiten.Image) {
	textFace := text.GoTextFace{
		Source: g.textFaceSource,
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

func (g *Game) drawArrow(screen *ebiten.Image) {
	angle := math.Atan2(g.pointerFront.y - g.pointerPivot.y, g.pointerFront.x - g.pointerPivot.x)

	drawGeom := ebiten.GeoM{}
	drawGeom.Translate(-g.pointerLength, -g.pointerWidth/2)
	drawGeom.Rotate(angle)
	drawGeom.Translate(g.pointerFront.x, g.pointerFront.y)
	screen.DrawImage(g.pointerImage, &ebiten.DrawImageOptions{GeoM: drawGeom})
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	menuItemFonts := g.getMenuItemFonts()
	menuItemBounds := g.getMenuItemBounds(menuItemFonts)
	for menuItemIndex, menuItem := range g.menuItems {
		menuItemBounds := menuItemBounds[menuItemIndex]
		drawGeom := ebiten.GeoM{}
		drawGeom.Translate(float64(menuItemBounds.Min.X), float64(menuItemBounds.Min.Y))

		colorScale := ebiten.ColorScale{}
		if menuItemIndex == g.selectedMenuItemIndex {
			colorScale.SetR(1)
			colorScale.SetG(1)
			colorScale.SetB(1)
		} else {
			colorScale.SetR(0.8)
			colorScale.SetG(0.8)
			colorScale.SetB(0.8)
		}

		drawOptions := &text.DrawOptions{
			DrawImageOptions: ebiten.DrawImageOptions{
				GeoM: drawGeom,
				ColorScale: colorScale,
			},
		}
		text.Draw(screen, menuItem.text, &menuItemFonts[menuItemIndex], drawOptions)
	}
}

func (g *Game) getMenuItemFonts() []text.GoTextFace {
	menuItemFonts := make([]text.GoTextFace, len(g.menuItems))
	for menuItemIndex, menuItem := range g.menuItems {
		var textZoomFactor float64
		isAnimating := !menuItem.animationEnd.IsZero() && menuItem.animationEnd.After(time.Now())
		if isAnimating {
			animationProgress := 1 - float64(time.Until(menuItem.animationEnd)) / float64(menuItemTextAnimationDuration)
			textZoomFactor = -animationProgress * (animationProgress - 2)
		} else {
			textZoomFactor = 1
			if !menuItem.animationEnd.IsZero() {
				menuItem.animationEnd = time.Time{}
			}
		}

		if menuItemIndex != g.selectedMenuItemIndex {
			textZoomFactor = 1 - textZoomFactor
		}

		textSize := textZoomFactor * selectedMenuItemTextSize + (1 - textZoomFactor) * unselectedMenuItemTextSize

		menuItemFonts[menuItemIndex] = text.GoTextFace{
			Source: g.textFaceSource,
			Size: textSize,
		}
	}
	return menuItemFonts
}

func (g *Game) getMenuItemBounds(menuItemFonts []text.GoTextFace) []image.Rectangle {
	menuItemsBounds := make([]image.Rectangle, len(g.menuItems))
	nextMenuItemTop := 0
	for menuItemIndex, menuItem := range g.menuItems {
		textWidth, textHeight := text.Measure(menuItem.text, &menuItemFonts[menuItemIndex], 0)
		menuItemsBounds[menuItemIndex] = image.Rect((g.screenSize.x - int(textWidth)) / 2, nextMenuItemTop, (g.screenSize.x + int(textWidth)) / 2, nextMenuItemTop + int(textHeight))
		nextMenuItemTop += int(textHeight)
	}

	menuHeight := nextMenuItemTop
	menuTop := (g.screenSize.y - menuHeight) / 2
	for menuItemIndex, menuItemBounds := range menuItemsBounds {
		menuItemsBounds[menuItemIndex] = image.Rect(menuItemBounds.Min.X, menuItemBounds.Min.Y + menuTop, menuItemBounds.Max.X, menuItemBounds.Max.Y + menuTop)
	}

	return menuItemsBounds
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

var menuItemTexts = []string{
	"Start game",
	"Exit",
}

var titleTextSize = float64(42)
var unselectedMenuItemTextSize = float64(28)
var selectedMenuItemTextSize = float64(42)

var menuItemTextAnimationDuration = time.Second * 1
