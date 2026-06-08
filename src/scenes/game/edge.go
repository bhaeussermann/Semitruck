package game

import "math"

type edge struct {
	centerLine *strokeLine
	width float64
}

func createEdge(x1 float64, y1 float64, x2 float64, y2 float64, width float64) *edge {
	return &edge{
		centerLine: &strokeLine{x1, y1, x2, y2},
		width: width,
	}
}

func (e *edge) getEdgeLines() []*edgeLine {
	x1 := e.centerLine.x1
	y1 := e.centerLine.y1
	x2 := e.centerLine.x2
	y2 := e.centerLine.y2
	perpendicularAngle := math.Atan2(x1 - x2, y2 - y1)
	dx := math.Cos(perpendicularAngle) * e.width / 2
	dy := math.Sin(perpendicularAngle) * e.width / 2

	return []*edgeLine {
		createEdgeLine(x1 - dx, y1 - dy, x2 - dx, y2 - dy),
		createEdgeLine(x1 + dx, y1 + dy, x2 + dx, y2 + dy),
	}
}

func (e *edge) getStartEdgeLine() *edgeLine {
	perpendicularAngle := math.Atan2(e.centerLine.x1 - e.centerLine.x2, e.centerLine.y2 - e.centerLine.y1)
	dx := math.Cos(perpendicularAngle) * e.width / 2
	dy := math.Sin(perpendicularAngle) * e.width / 2
	x1 := e.centerLine.x1
	y1 := e.centerLine.y1
	return createEdgeLine(x1 - dx, y1 - dy, x1 + dx, y1 + dy)
}

func (e *edge) getEndEdgeLine() *edgeLine {
	perpendicularAngle := math.Atan2(e.centerLine.x1 - e.centerLine.x2, e.centerLine.y2 - e.centerLine.y1)
	dx := math.Cos(perpendicularAngle) * e.width / 2
	dy := math.Sin(perpendicularAngle) * e.width / 2
	x2 := e.centerLine.x2
	y2 := e.centerLine.y2
	return createEdgeLine(x2 - dx, y2 - dy, x2 + dx, y2 + dy)
}

type strokeLine struct {
	x1, y1, x2, y2 float64
}

type edgeLine struct {
	m, c float64
	lowX, highX float64
}

func createEdgeLine(x1 float64, y1 float64, x2 float64, y2 float64) *edgeLine {
	if x1 == x2 {
		if y1 > y2 {
			swap(&y1, &y2)
		}
		return &edgeLine {
			m: math.MaxFloat64,
			c: x1,
			lowX: y1,
			highX: y2,
		}
	}

	if x1 > x2 {
		swap(&x1, &x2)
		swap(&y1, &y2)
	}

	m := (y2 - y1) / (x2 - x1)
	return &edgeLine{
		m: m,
		c: y1 - m * x1,
		lowX: x1,
		highX: x2,
	}
}

func (l *edgeLine) isHorizontal() bool {
	return l.m == 0
}

func (l *edgeLine) isVertical() bool {
	return l.m == math.MaxFloat64
}

func (l *edgeLine) getX(y float64) float64 {
	if l.isVertical() { return l.c }
	if l.m == 0 { return (l.lowX + l.highX) / 2 }
	return (y - l.c) / l.m
}

func (l *edgeLine) getY(x float64) float64 {
	if l.isVertical() { return 0 }
	return l.m * x + l.c
}

func (l *edgeLine) getDirection() float64 {
	if l.isVertical() { return math.Pi / 2 }
	return math.Atan(l.m)
}

func (l1 *edgeLine) intersectsLine(l2 *edgeLine) bool {
	if l1.isVertical() {
		if l2.isVertical() {
			return false
		}
		x := l1.c
		if (x < l2.lowX) || (x > l2.highX) { return false }
		y := l2.getY(x)
		return l1.lowX <= y && y <= l1.highX
	}

	if l2.isVertical() {
		x := l2.c
		if (x < l1.lowX) || (x > l1.highX) { return false }
		y := l1.getY(x)
		return l2.lowX <= y && y <= l2.highX
	}

	x := (l2.c - l1.c) / (l1.m - l2.m)
	return (l1.lowX <= x && x <= l1.highX) && (l2.lowX <= x && x <= l2.highX)
}

func (l1 *edgeLine) getIntersectPoint(l2 *edgeLine) point {
	if l1.isVertical() {
		x := l1.c
		return point {x, l2.getY(x)}
	}
	if l2.isVertical() {
		x := l2.c
		return point {x, l1.getY(x)}
	}
	x := (l2.c - l1.c) / (l1.m - l2.m)
	return point {x, l1.getY(x)}
}

func (l *edgeLine) arePointsOnSameSide(x1 float64, y1 float64, x2 float64, y2 float64) bool {
	if l.isVertical() {
		return (x1 < l.c) == (x2 < l.c)
	}
	return (y1 < l.getY(x1)) == (y2 < l.getY(x2))
}

func (l *edgeLine) getClosestPoint(x float64, y float64) (float64, float64, lineSection) {
	if l.isVertical() {
		if y <= l.lowX { return l.c, l.lowX, lineSection_Start }
		if y >= l.highX { return l.c, l.highX, lineSection_End }
		return l.c, y, lineSection_Middle
	}

	closestPointX := ((y - l.c) * l.m + x) / (l.m * l.m + 1)
	var lineSection lineSection = lineSection_Middle
	if closestPointX <= l.lowX {
		closestPointX = l.lowX
		lineSection = lineSection_Start
	} else if closestPointX >= l.highX {
		closestPointX = l.highX
		lineSection = lineSection_End
	}
	closestPointY := l.getY(closestPointX)
	return closestPointX, closestPointY, lineSection
}

type lineSection byte

const (
	lineSection_Start = iota
	lineSection_Middle
	lineSection_End
)

func (l *edgeLine) getStrokeLine() *strokeLine {
	if l.isVertical() {
		return &strokeLine{l.c, l.lowX, l.c, l.highX}
	}
	return &strokeLine{l.lowX, l.getY(l.lowX), l.highX, l.getY(l.highX)}
}

func swap(x1 *float64, x2 *float64) {
	t := *x1
	*x1 = *x2
	*x2 = t
}
