package game

type border struct {
	edges []*edge
	width float64
}

func createBorder(points []point, width float64) *border {
	edges := make([]*edge, len(points) - 1)
	for pointIndex := 0; pointIndex < len(points) - 1; pointIndex++ {
		currentPoint := points[pointIndex]
		nextPoint := points[pointIndex + 1]
		edges[pointIndex] = createEdge(currentPoint.x, currentPoint.y, nextPoint.x, nextPoint.y, width)
	}
	return &border {edges, width}
}

type point struct {
	x, y float64
}
