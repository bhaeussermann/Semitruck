package game

type semitrailer struct {
	spriteWidth, spriteLength float64
	width, length float64
	kingPinLocation float64
	frontX, frontY float64
	direction float64
}

func (t *semitrailer) getDimensions() *dimensions {
	return &dimensions{t.width, t.length}
}

func (t *semitrailer) getLocation() *location {
	return &location{t.frontX, t.frontY, t.direction}
}
