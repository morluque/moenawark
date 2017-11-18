package universe

import "math"

type point struct {
	x float64
	y float64
}

func (p point) equal(b point) bool {
	return p.x == b.x && p.y == b.y
}

func dist(p1, p2 point) float64 {
	return math.Sqrt((p1.x-p2.x)*(p1.x-p2.x) + (p1.y-p2.y)*(p1.y-p2.y))
}

// EPSILON is the precision to use for cross-product comparison
const EPSILON float64 = 0.000001

func crossProduct(a, b point) float64 {
	// The mathematical definition applies to axis going north east, but ours go
	// south east; so we work with the opposite of y coordinate
	// Doesn't really matter as all we're really interested in is wether cross
	// product is the same for two segments starting from origin.
	return a.x*-b.y - b.x*-a.y
}

func (p point) translate(a point) point {
	return point{x: p.x - a.x, y: p.y - a.y}
}

func (p point) farEnough(minDist float64, points ...point) bool {
	for _, p2 := range points {
		if p.equal(p2) || dist(p, p2) <= minDist {
			return false
		}
	}
	return true
}

type segment struct {
	a point
	b point
}

func (s segment) equal(s2 segment) bool {
	return s.a.equal(s2.a) && s.b.equal(s2.b)
}

func (s segment) translate(a point) segment {
	return segment{a: point{x: s.a.x - a.x, y: s.a.y - a.y}, b: point{x: s.b.x - a.x, y: s.b.y - a.y}}
}

func (s segment) boundingBox() segment {
	a := point{x: s.a.x, y: s.a.y}
	b := point{x: s.b.x, y: s.b.y}
	// Axis extend to south east
	if a.x > b.x {
		// if a east of b, swap x
		a.x = s.b.x
		b.x = s.a.x
	}
	if a.y > b.y {
		// if a south of b, swap y
		a.y = s.b.y
		b.y = s.a.y
	}

	return segment{a: a, b: b}
}

func doBBoxIntersect(s1, s2 segment) bool {
	return (s1.a.x < s2.b.x) && (s1.b.x > s2.a.x) && (s1.a.y < s2.b.y) && (s1.b.y > s2.a.y)
}

func isPointOnSegment(s segment, p point) bool {
	if s.a.equal(p) || s.b.equal(p) {
		return false
	}
	// Translate s and p so that s.a is at origin
	sTransOrig := s.translate(s.a)
	pTransOrig := p.translate(s.a)
	cp := crossProduct(sTransOrig.b, pTransOrig)
	return math.Abs(cp) < EPSILON
}

func isPointRightOfLine(s segment, p point) bool {
	// Translate s and p so that s.a is at origin
	sTransOrig := s.translate(s.a)
	pTransOrig := p.translate(s.a)
	return crossProduct(sTransOrig.b, pTransOrig) > 0
}

func doesSegmentCrosses(s1, s2 segment) bool {
	if isPointOnSegment(s1, s2.a) || isPointOnSegment(s1, s2.b) {
		return true
	}
	return isPointRightOfLine(s1, s2.a) != isPointRightOfLine(s1, s2.b)
}

func segmentIntersect(s1, s2 segment) bool {
	if s1.a.equal(s2.a) || s1.a.equal(s2.b) || s1.b.equal(s2.a) || s1.b.equal(s2.b) {
		return false
	}
	b1 := s1.boundingBox()
	b2 := s2.boundingBox()
	if !doBBoxIntersect(b1, b2) {
		return false
	}
	return doesSegmentCrosses(s1, s2) && doesSegmentCrosses(s2, s1)
}

func copySegment(s segment) segment {
	a := point{x: s.a.x, y: s.a.y}
	b := point{x: s.b.x, y: s.b.y}

	return segment{a: a, b: b}
}

func (s segment) intersect(segments ...segment) bool {
	for _, s2 := range segments {
		if segmentIntersect(s, s2) {
			return true
		}
	}
	return false
}
