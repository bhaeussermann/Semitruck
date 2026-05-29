package game

import "math"

type edge struct {
	centerLine *strokeLine
	edgeLine1, edgeLine2 *edgeLine
	width float64
}

func createEdge(x1 float64, y1 float64, x2 float64, y2 float64, width float64) *edge {
	perpendicularAngle := math.Atan2(x1 - x2, y2 - y1)
	dx := math.Cos(perpendicularAngle) * width / 2
	dy := math.Sin(perpendicularAngle) * width / 2

	return &edge{
		centerLine: &strokeLine{x1, y1, x2, y2},
		edgeLine1: createEdgeLine(x1 - dx, y1 - dy, x2 - dx, y2 - dy),
		edgeLine2: createEdgeLine(x1 + dx, y1 + dy, x2 + dx, y2 + dy),
		width: width,
	}
}

func (line *edgeLine) intersectsEdge(edge *edge) bool {
	return line.intersectsLine(edge.edgeLine1) || line.intersectsLine(edge.edgeLine2)
}

func (l1 *edgeLine) intersectsLine(l2 *edgeLine) bool {
	if l1.isVertical() {
		if l2.isVertical() {
			return false
		}
		x := l1.c
		if (x < l2.lowX) || (x > l2.highX) { return false }
		y := l2.m * x + l2.c
		return l1.lowX <= y && y <= l1.highX
	}

	if l2.isVertical() {
		x := l2.c
		if (x < l1.lowX) || (x > l1.highX) { return false }
		y := l1.m * x + l1.c
		return l2.lowX <= y && y <= l2.highX
	}

	x := (l2.c - l1.c) / (l1.m - l2.m)
	return (l1.lowX <= x && x <= l1.highX) && (l2.lowX <= x && x <= l2.highX)
}

func (edge *edge) getClosestLine(x float64, y float64) *edgeLine {
	if edge.edgeLine1.isVertical() {
		if math.Abs(x - edge.edgeLine1.c) <= math.Abs(x - edge.edgeLine2.c) {
			return edge.edgeLine1
		} else {
			return edge.edgeLine2
		}
	}

	if math.Abs(y - edge.edgeLine1.getY(x)) <= math.Abs(y - edge.edgeLine2.getY(x)) {
		return edge.edgeLine1
	} else {
		return edge.edgeLine2
	}
}

type strokeLine struct {
	x1, y1, x2, y2 float64
}

func (line *strokeLine) getDirection() float64 {
	x1 := line.x1
	y1 := line.y1
	x2 := line.x2
	y2 := line.y2

	if y1 == y2 { return 0 }

	if y2 < y1 {
		swap(&x1, &x2)
		swap(&y1, &y2)
	}

	return math.Atan2(y2 - y1, x2 - x1)
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

func (l *edgeLine) isVertical() bool {
	return l.m == math.MaxFloat64
}

func (l *edgeLine) getY(x float64) float64 {
	if l.isVertical() { return 0 }
	return l.m * x + l.c
}

func (l *edgeLine) arePointsOnSameSide(x1 float64, y1 float64, x2 float64, y2 float64) bool {
	if l.isVertical() {
		return (x1 < l.c) == (x2 < l.c)
	}
	return (y1 < l.getY(x1)) == (y2 < l.getY(x2))
}

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
