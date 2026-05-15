package menu

import (
	"bytes"
	"image"
	"math"
	"time"

	"github.com/bhaeussermann/semitruck/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gobold"
)

type Menu struct {
	screenSize components.Coordinates
	pointerFront components.Coordinatesf
	pointerPivot components.Coordinatesf
	displayPointer bool
	selectedMenuItemIndex MenuItemSelectionIndex
	menuItems []*menuItem

	pointerWidth float64
	pointerLength float64
	pointerImage *ebiten.Image
	textFaceSource *text.GoTextFaceSource
}

type menuItem struct {
	text string
	animationEnd time.Time
}

type MenuItemSelectionIndex int

var MenuItemSelectionIndex_None MenuItemSelectionIndex = -1

func NewMenu(menuItemTexts []string) (*Menu, error) {
	pointerImage, _, error := ebitenutil.NewImageFromFile("../images/pointer.png")
	if error != nil {
		return nil, error
	}

	textFaceSource, error := text.NewGoTextFaceSource(bytes.NewReader(gobold.TTF))
	if error != nil {
		return nil, error
	}

	menuItems := make([]*menuItem, len(menuItemTexts))
	for menuItemIndex, menuItemText := range menuItemTexts {
		menuItems[menuItemIndex] = &menuItem{ text: menuItemText, }
	}

	cursorX, cursorY := ebiten.CursorPosition()
	return &Menu{
		pointerFront: components.Coordinatesf{X: float64(cursorX), Y: float64(cursorY)},
		pointerPivot: components.NilCoordinatesf,
		selectedMenuItemIndex: MenuItemSelectionIndex_None,
		menuItems: menuItems,

		pointerWidth: float64(pointerImage.Bounds().Dy()),
		pointerLength: float64(pointerImage.Bounds().Dx()),
		pointerImage: pointerImage,
		textFaceSource: textFaceSource,
	}, nil
}

func (m *Menu) SetScreenSize(width int, height int) {
	m.screenSize = components.Coordinates{X: width, Y: height}
}

func (m * Menu) Update() MenuItemSelectionIndex {
	if (m.selectedMenuItemIndex != MenuItemSelectionIndex_None) && (ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || ebiten.IsKeyPressed(ebiten.KeyEnter)) {
		return m.selectedMenuItemIndex
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		m.updateMenuItemSelectionFromKeyPress()
		m.displayPointer = false
	} else {
		cursorX, cursorY := ebiten.CursorPosition()
		if (cursorX >= 0) && (cursorY >= 0) && (cursorX < m.screenSize.X) && (cursorY < m.screenSize.Y) {
			if (cursorX != int(m.pointerFront.X)) || (cursorY != int(m.pointerFront.Y)) {
				m.updateCursorPointer(cursorX, cursorY)
				m.setMenuItemSelectionFromCursor(cursorX, cursorY)
			}
		} else {
			m.pointerFront = components.NilCoordinatesf
			m.displayPointer = false
		}
	}

	return -1
}

func (m *Menu) updateCursorPointer(cursorX int, cursorY int) {
	m.pointerFront = components.Coordinatesf{X: float64(cursorX), Y: float64(cursorY)}

	pointerAngle := math.Atan2(m.pointerFront.Y - m.pointerPivot.Y, m.pointerFront.X - m.pointerPivot.X)
	pointerPivotLocation := m.pointerLength * pointerPivotLengthRatio
	m.pointerPivot = components.Coordinatesf{
		X: m.pointerFront.X - pointerPivotLocation * math.Cos(pointerAngle),
		Y: m.pointerFront.Y - pointerPivotLocation * math.Sin(pointerAngle),
	}
	m.displayPointer = true
}

func (m *Menu) updateMenuItemSelectionFromKeyPress() {
	var nextSelectedMenuItemIndex MenuItemSelectionIndex
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		nextSelectedMenuItemIndex = m.selectedMenuItemIndex - 1
		if nextSelectedMenuItemIndex < 0 {
			nextSelectedMenuItemIndex = MenuItemSelectionIndex(len(m.menuItems) - 1)
		}
	} else {
		nextSelectedMenuItemIndex = m.selectedMenuItemIndex + 1
		if nextSelectedMenuItemIndex >= MenuItemSelectionIndex(len(m.menuItems)) {
			nextSelectedMenuItemIndex = 0
		}
	}
	m.setMenuItemSelection(nextSelectedMenuItemIndex)
}

func (m *Menu) setMenuItemSelectionFromCursor(cursorX int, cursorY int) {
	menuItemSelectionIndex := m.getMenuItemSelectionFromCursor(cursorX, cursorY)
	m.setMenuItemSelection(menuItemSelectionIndex)
}

func (m *Menu) getMenuItemSelectionFromCursor(cursorX int, cursorY int) MenuItemSelectionIndex {
	menuItemsBounds := m.getMenuItemBounds(m.getMenuItemFonts())
	for menuItemIndex, menuItemBounds := range menuItemsBounds {
		if (cursorX >= menuItemBounds.Min.X) && (cursorX <= menuItemBounds.Max.X) && (cursorY >= menuItemBounds.Min.Y) && (cursorY <= menuItemBounds.Max.Y) {
			return MenuItemSelectionIndex(menuItemIndex)
		}
	}
	return MenuItemSelectionIndex_None
}

func (m *Menu) setMenuItemSelection(nextSelectedMenuItemIndex MenuItemSelectionIndex) {
	if nextSelectedMenuItemIndex != m.selectedMenuItemIndex {
		if m.selectedMenuItemIndex != MenuItemSelectionIndex_None {
			m.updateMenuItemAnimation(m.selectedMenuItemIndex)
		}
		m.selectedMenuItemIndex = nextSelectedMenuItemIndex
		if m.selectedMenuItemIndex != MenuItemSelectionIndex_None {
			m.updateMenuItemAnimation(m.selectedMenuItemIndex)
		}
	}
}

func (m *Menu) updateMenuItemAnimation(menuItemIndex MenuItemSelectionIndex) {
	menuItem := m.menuItems[menuItemIndex]
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

func (m *Menu) Draw(screen *ebiten.Image) {
	m.drawMenu(screen)
	if m.displayPointer {
		m.drawPointer(screen)
	}
}

func (m *Menu) drawPointer(screen *ebiten.Image) {
	angle := math.Atan2(m.pointerFront.Y - m.pointerPivot.Y, m.pointerFront.X - m.pointerPivot.X)

	drawGeom := ebiten.GeoM{}
	drawGeom.Translate(-m.pointerLength, -m.pointerWidth / 2)
	drawGeom.Rotate(angle)
	drawGeom.Translate(m.pointerFront.X, m.pointerFront.Y)
	screen.DrawImage(m.pointerImage, &ebiten.DrawImageOptions{GeoM: drawGeom})
}

func (m *Menu) drawMenu(screen *ebiten.Image) {
	menuItemFonts := m.getMenuItemFonts()
	menuItemBounds := m.getMenuItemBounds(menuItemFonts)
	for menuItemIndex, menuItem := range m.menuItems {
		menuItemBounds := menuItemBounds[menuItemIndex]
		drawGeom := ebiten.GeoM{}
		drawGeom.Translate(float64(menuItemBounds.Min.X), float64(menuItemBounds.Min.Y))

		colorScale := ebiten.ColorScale{}
		if MenuItemSelectionIndex(menuItemIndex) == m.selectedMenuItemIndex {
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

func (m *Menu) getMenuItemFonts() []text.GoTextFace {
	menuItemFonts := make([]text.GoTextFace, len(m.menuItems))
	for menuItemIndex, menuItem := range m.menuItems {
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

		if MenuItemSelectionIndex(menuItemIndex) != m.selectedMenuItemIndex {
			textZoomFactor = 1 - textZoomFactor
		}

		textSize := textZoomFactor * selectedMenuItemTextSize + (1 - textZoomFactor) * unselectedMenuItemTextSize

		menuItemFonts[menuItemIndex] = text.GoTextFace{
			Source: m.textFaceSource,
			Size: textSize,
		}
	}
	return menuItemFonts
}

func (m *Menu) getMenuItemBounds(menuItemFonts []text.GoTextFace) []image.Rectangle {
	menuItemsBounds := make([]image.Rectangle, len(m.menuItems))
	nextMenuItemTop := 0
	for menuItemIndex, menuItem := range m.menuItems {
		textWidth, textHeight := text.Measure(menuItem.text, &menuItemFonts[menuItemIndex], 0)
		menuItemsBounds[menuItemIndex] = image.Rect((m.screenSize.X - int(textWidth)) / 2, nextMenuItemTop, (m.screenSize.X + int(textWidth)) / 2, nextMenuItemTop + int(textHeight))
		nextMenuItemTop += int(textHeight)
	}

	menuHeight := nextMenuItemTop
	menuTop := (m.screenSize.Y - menuHeight) / 2
	for menuItemIndex, menuItemBounds := range menuItemsBounds {
		menuItemsBounds[menuItemIndex] = image.Rect(menuItemBounds.Min.X, menuItemBounds.Min.Y + menuTop, menuItemBounds.Max.X, menuItemBounds.Max.Y + menuTop)
	}

	return menuItemsBounds
}

var pointerPivotLengthRatio = float64(0.5)

var unselectedMenuItemTextSize = float64(28)
var selectedMenuItemTextSize = float64(42)

var menuItemTextAnimationDuration = time.Second * 1
