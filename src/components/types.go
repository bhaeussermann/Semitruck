package components

type Coordinates struct {
	X int
	Y int
}

type Coordinatesf struct {
	X float64
	Y float64
}

var NilCoordinatesf = Coordinatesf{X: -1, Y: -1}
