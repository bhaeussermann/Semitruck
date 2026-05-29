package game

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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
	bumpVelocityX float64
	bumpVelocityY float64
}

func (t *truck) updateMovement() {
	t.updateSpeed()
	t.turn()
	t.moveFromBumpVelocity()
	t.propel()
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

	t.bumpVelocityX *= bumpFrictionFactor
	t.bumpVelocityY *= bumpFrictionFactor
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
}

func (t *truck) moveFromBumpVelocity() {
	if math.Abs(t.bumpVelocityX) + math.Abs(t.bumpVelocityY) <= epsilon {
		return
	}

	rearWheelX := t.frontX - t.length * rearWheelLengthRatio * math.Cos(t.direction)
	rearWheelY := t.frontY - t.length * rearWheelLengthRatio * math.Sin(t.direction)

	t.frontX += t.bumpVelocityX
	t.frontY += t.bumpVelocityY
	t.direction = math.Atan2(t.frontY - rearWheelY, t.frontX - rearWheelX)
}

func (t *truck) propel() {
	if math.Abs(t.speed) <= epsilon {
		return
	}

	rearWheelX := t.frontX - t.length * rearWheelLengthRatio * math.Cos(t.direction)
	rearWheelY := t.frontY - t.length * rearWheelLengthRatio * math.Sin(t.direction)

	if math.Abs(t.wheelTurnDirection) > epsilon {
		effectiveTurnAngle := math.Asin(t.speed * math.Sin(t.wheelTurnDirection*2) / math.Sqrt(t.speed*t.speed + t.wheelDistance*t.wheelDistance + 2 * t.speed * t.length * math.Cos(t.wheelTurnDirection*2)))
		t.direction += effectiveTurnAngle
		if t.direction > math.Pi {
			t.direction -= 2 * math.Pi
		} else if t.direction < -math.Pi {
			t.direction += 2 * math.Pi
		}
	}

	rearWheelX += t.speed * math.Cos(t.direction)
	rearWheelY += t.speed * math.Sin(t.direction)
	t.frontX = rearWheelX + t.length * rearWheelLengthRatio * math.Cos(t.direction)
	t.frontY = rearWheelY + t.length * rearWheelLengthRatio * math.Sin(t.direction)
}

func (t *truck) bump(truckEdge *edgeLine, courseEdge *edge) {
	truckVelocityX := t.speed * math.Cos(t.direction) + t.bumpVelocityX
	truckVelocityY := t.speed * math.Sin(t.direction) + t.bumpVelocityY
	truckSpeed := math.Sqrt(truckVelocityX*truckVelocityX + truckVelocityY*truckVelocityY)
	truckMovementDirection := math.Atan2(truckVelocityY, truckVelocityX)
	if truckMovementDirection < 0 {
		truckMovementDirection += 2 * math.Pi
	}

	courseEdgeDirection := courseEdge.centerLine.getDirection()
	opposingForceMagnitude := bumpForceCoefficient * truckSpeed * math.Abs(math.Sin(courseEdgeDirection - truckMovementDirection))
	var opposingForceDirection float64
	if courseEdgeDirection < truckMovementDirection && truckMovementDirection <= courseEdgeDirection + math.Pi {
		opposingForceDirection = courseEdgeDirection - math.Pi / 2
	} else {
		opposingForceDirection = courseEdgeDirection + math.Pi / 2
	}

	t.bumpVelocityX = truckVelocityX + opposingForceMagnitude * math.Cos(opposingForceDirection)
	t.bumpVelocityY = truckVelocityY + opposingForceMagnitude * math.Sin(opposingForceDirection)
	t.speed = 0

	t.pushBack(truckEdge, courseEdge)
}

func (t *truck) pushBack(truckEdge *edgeLine, courseEdge *edge) {
	centerX, centerY := t.getCenter()
	courseLine := courseEdge.getClosestLine(centerX, centerY)
	truckEdgeEndpoints := truckEdge.getStrokeLine()

	var oppositeX float64
	var oppositeY float64
	if !courseLine.arePointsOnSameSide(centerX, centerY, truckEdgeEndpoints.x1, truckEdgeEndpoints.y1) {
		oppositeX = truckEdgeEndpoints.x1
		oppositeY = truckEdgeEndpoints.y1
	} else if !courseLine.arePointsOnSameSide(centerX, centerY, truckEdgeEndpoints.x2, truckEdgeEndpoints.y2) {
		oppositeX = truckEdgeEndpoints.x2
		oppositeY = truckEdgeEndpoints.y2
	} else {
		return
	}

	if courseLine.isVertical() {
		t.frontX -= oppositeX - courseLine.c
		return
	}

	courseLineIntersectX := ((oppositeY - courseLine.c) * courseLine.m + oppositeX) / (courseLine.m * courseLine.m + 1)
	courseLineIntersectY := courseLine.getY(courseLineIntersectX)
	pushDistance := math.Sqrt(sqr(courseLineIntersectY - oppositeY) + sqr(courseLineIntersectX - oppositeX))
	pushAngle := math.Atan2(courseLineIntersectY - oppositeY, courseLineIntersectX - oppositeX)
	t.frontX += pushDistance * math.Cos(pushAngle)
	t.frontY += pushDistance * math.Sin(pushAngle)
}

func sqr(x float64) float64 { return x * x }

func (t *truck) getCenter() (x float64, y float64) {
	return t.frontX - math.Cos(t.direction) * t.length / 2, t.frontY - math.Sin(t.direction) * t.length / 2
}

var acceleration = 0.03
var frictionFactor = 0.95
var breakFactor = 0.9
var bumpForceCoefficient = 1.5
var bumpFrictionFactor = 0.95
var maximumForwardSpeed = 4.0
var maximumReverseSpeed = 1.5
var maxTurnAngle = math.Pi / 4
var turnSpeed = maxTurnAngle / 10

var epsilon = 0.00001
