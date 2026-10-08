package game

import "math"

type vehicle interface {
	getDimensions() *dimensions
	getLocation() *location
}

type dimensions struct {
	Width, Length float64
}

type location struct {
	FrontX, FrontY float64
	Direction float64
}

func getEdges(vehicle vehicle) []*edgeLine {
	dimensions := vehicle.getDimensions()
	location := vehicle.getLocation()

	frontLeftX := location.FrontX + math.Sin(location.Direction) * dimensions.Width / 2
	frontLeftY := location.FrontY - math.Cos(location.Direction) * dimensions.Width / 2
	rearLeftX := frontLeftX - math.Cos(location.Direction) * dimensions.Length
	rearLeftY := frontLeftY - math.Sin(location.Direction) * dimensions.Length
	frontRightX := location.FrontX - math.Sin(location.Direction) * dimensions.Width / 2
	frontRightY := location.FrontY + math.Cos(location.Direction) * dimensions.Width / 2
	rearRightX := frontRightX - math.Cos(location.Direction) * dimensions.Length
	rearRightY := frontRightY - math.Sin(location.Direction) * dimensions.Length

	return []*edgeLine{
		createEdgeLine(frontLeftX, frontLeftY, frontRightX, frontRightY, location.Direction),
		createEdgeLine(rearLeftX, rearLeftY, rearRightX, rearRightY, location.Direction + math.Pi),
		createEdgeLine(frontLeftX, frontLeftY, rearLeftX, rearLeftY, location.Direction - math.Pi / 2),
		createEdgeLine(frontRightX, frontRightY, rearRightX, rearRightY, location.Direction + math.Pi / 2),
	}
}
