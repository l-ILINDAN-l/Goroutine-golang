package main

import (
	"fmt"
	"math"
)

type Point struct {
	x, y float64
}

func NewPoint(x, y float64) Point {
	return Point{x, y}
}

func (p Point) Distance(other Point) float64 {
	return math.Sqrt(math.Pow(p.y-other.y, 2) + math.Pow(p.x-other.x, 2))
}

func main() {
	p1 := NewPoint(0.0, 1.0)
	p2 := NewPoint(3.0, 5.0)

	fmt.Println(p1.Distance(p2))
}
