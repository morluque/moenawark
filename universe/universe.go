package universe

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
)

// Universe stores places and ways between them
type Universe struct {
	Width        int
	Height       int
	MinPlaceDist float64
	MaxWayLength float64
	RegionCount  int
	RegionRadius float64
	Places       []Place
	Ways         []Way
	points       []Point
	segments     []Segment
	dists        map[Segment]float64
}

func newUniverse(width, height int, minPlaceDist, maxWayLength float64) *Universe {
	u := Universe{
		Width:        width,
		Height:       height,
		MinPlaceDist: minPlaceDist,
		MaxWayLength: maxWayLength,
	}
	u.points = make([]Point, 0)
	u.segments = make([]Segment, 0)
	u.dists = make(map[Segment]float64)

	return &u
}

// Point in 2D
type Point struct {
	X float64
	Y float64
}

// Segment in 2D
type Segment struct {
	A Point
	B Point
}

// Place in the universe
type Place struct {
	ID   int64
	P    Point
	Name string
}

// Way between two places
type Way struct {
	ID  int64
	Src Place
	Dst Place
	Len float64
}

func (p Point) equal(b Point) bool {
	return p.X == b.X && p.Y == b.Y
}

func (s Segment) equal(s2 Segment) bool {
	return s.A.equal(s2.A) && s.B.equal(s2.B)
}

// WriteDotFile exports a universe to Graphviz "dot" format
func (u *Universe) WriteDotFile(path string) error {
	var scale float64 = 20
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("digraph G {\n")
	if err != nil {
		return err
	}
	for _, p := range u.Places {
		_, err = f.WriteString(fmt.Sprintf("    p%d [pos=\"%f,%f!\"];\n", p.ID, p.P.X*scale, p.P.Y*scale))
		if err != nil {
			return err
		}
	}
	for _, w := range u.Ways {
		_, err = f.WriteString(fmt.Sprintf("    p%d -> p%d;\n", w.Src.ID, w.Dst.ID))
		if err != nil {
			return err
		}
	}
	_, err = f.WriteString("}\n")
	return err
}

var nextPlaceID int64 = 1

func newPlace(p Point) Place {
	place := Place{ID: nextPlaceID, P: p, Name: "TODO"}
	nextPlaceID++
	return place
}

var nextWayID int64 = 1

func newWay(s Segment, places []Place) Way {
	src := Place{ID: 0}
	dst := Place{ID: 0}
	for _, pl := range places {
		if pl.P.equal(s.A) {
			src.ID = pl.ID
			src.P.X = pl.P.X
			src.P.Y = pl.P.Y
			src.Name = pl.Name
		}
		if pl.P.equal(s.B) {
			dst.ID = pl.ID
			dst.P.X = pl.P.X
			dst.P.Y = pl.P.Y
			dst.Name = pl.Name
		}
		if src.ID != 0 && dst.ID != 0 {
			break
		}
	}
	w := Way{ID: nextWayID, Src: src, Dst: dst, Len: dist(src.P, dst.P)}
	nextWayID++

	return w
}

// EPSILON is the precision to use for cross-product comparison
const EPSILON float64 = 0.000001

func dist(p1, p2 Point) float64 {
	return math.Sqrt((p1.X-p2.X)*(p1.X-p2.X) + (p1.Y-p2.Y)*(p1.Y-p2.Y))
}

func crossProduct(a, b Point) float64 {
	// The mathematical definition applies to axis going north east, but ours go
	// south east; so we work with the opposite of Y coordinate
	// Doesn't really matter as all we're really interested in is wether cross
	// product is the same for two segments starting from origin.
	return a.X*-b.Y - b.X*-a.Y
}

func (p Point) translate(a Point) Point {
	return Point{X: p.X - a.X, Y: p.Y - a.Y}
}

func (s Segment) translate(a Point) Segment {
	return Segment{A: Point{X: s.A.X - a.X, Y: s.A.Y - a.Y}, B: Point{X: s.B.X - a.X, Y: s.B.Y - a.Y}}
}

func (s Segment) boundingBox() Segment {
	a := Point{X: s.A.X, Y: s.A.Y}
	b := Point{X: s.B.X, Y: s.B.Y}
	// Axis extend to south east
	if a.X > b.X {
		// if a east of b, swap X
		a.X = s.B.X
		b.X = s.A.X
	}
	if a.Y > b.Y {
		// if a south of b, swap Y
		a.Y = s.B.Y
		b.Y = s.A.Y
	}

	return Segment{A: a, B: b}
}

func doBBoxIntersect(s1, s2 Segment) bool {
	return (s1.A.X < s2.B.X) && (s1.B.X > s2.A.X) && (s1.A.Y < s2.B.Y) && (s1.B.Y > s2.A.Y)
}

func isPointOnSegment(s Segment, p Point) bool {
	if s.A.equal(p) || s.B.equal(p) {
		return false
	}
	// Translate s and p so that s.A is at origin
	sTransOrig := s.translate(s.A)
	pTransOrig := p.translate(s.A)
	cp := crossProduct(sTransOrig.B, pTransOrig)
	return math.Abs(cp) < EPSILON
}

func isPointRightOfLine(s Segment, p Point) bool {
	// Translate s and p so that s.A is at origin
	sTransOrig := s.translate(s.A)
	pTransOrig := p.translate(s.A)
	return crossProduct(sTransOrig.B, pTransOrig) > 0
}

func doesSegmentCrosses(s1, s2 Segment) bool {
	if isPointOnSegment(s1, s2.A) || isPointOnSegment(s1, s2.B) {
		return true
	}
	return isPointRightOfLine(s1, s2.A) != isPointRightOfLine(s1, s2.B)
}

func segmentIntersect(s1, s2 Segment) bool {
	if s1.A.equal(s2.A) || s1.A.equal(s2.B) || s1.B.equal(s2.A) || s1.B.equal(s2.B) {
		return false
	}
	b1 := s1.boundingBox()
	b2 := s2.boundingBox()
	if !doBBoxIntersect(b1, b2) {
		return false
	}
	return doesSegmentCrosses(s1, s2) && doesSegmentCrosses(s2, s1)
}

func (u *Universe) generatePoints() {
	fail := 0
	for {
		fail++
		if fail > (u.Width+u.Height)/2*100 {
			break
		}
		newp := Point{X: rand.Float64() * float64(u.Width), Y: rand.Float64() * float64(u.Height)}
		ok := true
		for _, p := range u.points {
			if p.equal(newp) || dist(p, newp) <= u.MinPlaceDist {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		fail = 0
		u.points = append(u.points, newp)
	}
	log.Printf("%d points generated\n", len(u.points))
}

func (u *Universe) computeDists() {
	for _, a := range u.points {
		for _, b := range u.points {
			if a.equal(b) {
				continue
			}
			if d := dist(a, b); d <= u.MaxWayLength {
				u.dists[Segment{A: a, B: b}] = d
				u.dists[Segment{A: b, B: a}] = d
			}
		}
	}
	log.Printf("%d potential segments\n", len(u.dists))
}

func copySegment(s Segment) Segment {
	a := Point{X: s.A.X, Y: s.A.Y}
	b := Point{X: s.B.X, Y: s.B.Y}

	return Segment{A: a, B: b}
}

func (u *Universe) generateSegments() {
	for _, a := range u.points {
		for news := range u.dists {
			if news.A.equal(a) {
				ok := true
				for _, s := range u.segments {
					if segmentIntersect(s, news) {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}
				u.segments = append(u.segments, copySegment(news))
			}
		}
	}
	log.Printf("%d segments generated\n", len(u.segments))
}

func (u *Universe) makePlacesAndWays() {
	u.Places = make([]Place, len(u.points))
	for i, p := range u.points {
		u.Places[i] = newPlace(p)
	}

	u.Ways = make([]Way, len(u.segments))
	for i, s := range u.segments {
		u.Ways[i] = newWay(s, u.Places)
	}
}

func (u *Universe) cleanup() {
	u.points = make([]Point, 0)
	u.segments = make([]Segment, 0)
	u.dists = make(map[Segment]float64)
}

// Generate generates a new random universe
func Generate(width, height int, minDist, maxWayLength float64) *Universe {
	u := newUniverse(width, height, minDist, maxWayLength)

	log.Printf("Generating places at least distant from %d... ", int(minDist))
	u.generatePoints()
	log.Printf("OK\n")

	log.Printf("Computing distances beneath %d... ", int(maxWayLength))
	u.computeDists()
	log.Printf("OK\n")

	log.Printf("Computing ways... ")
	u.generateSegments()
	log.Printf("OK\n")

	u.makePlacesAndWays()

	u.cleanup()

	return u
}
