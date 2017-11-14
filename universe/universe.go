package universe

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
)

type RegionConfig struct {
	Count        int
	Radius       float64
	MinPlaceDist float64
	MaxWayLength float64
}

type Config struct {
	Radius       float64
	MinPlaceDist float64
	MaxWayLength float64
	RegionConfig RegionConfig
}

// Universe stores places and ways between them
type Universe struct {
	Config
	Region  *Region
	Regions []*Region
	Places  []Place
	Ways    []Way
}

func newUniverse(cfg Config) *Universe {
	u := Universe{
		Config: cfg,
	}
	u.Region = newRegion(point{x: u.Radius, y: u.Radius}, u.Radius)

	return &u
}

// point in 2D
type point struct {
	x float64
	y float64
}

type Region struct {
	Center   point
	Radius   float64
	points   []point
	segments []segment
	dists    map[segment]float64
}

func newRegion(center point, radius float64) *Region {
	r := Region{
		Center: point{x: center.x, y: center.y},
		Radius: radius,
	}

	r.points = make([]point, 0)
	r.segments = make([]segment, 0)
	r.dists = make(map[segment]float64)

	return &r
}

// segment in 2D
type segment struct {
	a point
	b point
}

// Place in the universe
type Place struct {
	ID   int64
	P    point
	Name string
}

// Way between two places
type Way struct {
	ID  int64
	Src Place
	Dst Place
	Len float64
}

func (p point) equal(b point) bool {
	return p.x == b.x && p.y == b.y
}

func (s segment) equal(s2 segment) bool {
	return s.a.equal(s2.a) && s.b.equal(s2.b)
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
		_, err = f.WriteString(fmt.Sprintf("    p%d [pos=\"%f,%f!\"];\n", p.ID, p.P.x*scale, p.P.y*scale))
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

func newPlace(p point) Place {
	place := Place{ID: nextPlaceID, P: p, Name: "TODO"}
	nextPlaceID++
	return place
}

var nextWayID int64 = 1

func newWay(s segment, places []Place) Way {
	src := Place{ID: 0}
	dst := Place{ID: 0}
	for _, pl := range places {
		if pl.P.equal(s.a) {
			src.ID = pl.ID
			src.P.x = pl.P.x
			src.P.y = pl.P.y
			src.Name = pl.Name
		}
		if pl.P.equal(s.b) {
			dst.ID = pl.ID
			dst.P.x = pl.P.x
			dst.P.y = pl.P.y
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

func dist(p1, p2 point) float64 {
	return math.Sqrt((p1.x-p2.x)*(p1.x-p2.x) + (p1.y-p2.y)*(p1.y-p2.y))
}

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

func randPointInRegion(r *Region) point {
	for i := 0; i < 1000; i++ {
		p := point{x: rand.Float64() * r.Radius * 2, y: rand.Float64() * r.Radius * 2}
		if dist(r.Center, p) <= r.Radius {
			return p
		}
	}
	return point{}
}

func (r *Region) generatePoints(minPlaceDist float64) {
	fail := 0
	for {
		fail++
		if fail > int(r.Radius)*100 {
			break
		}
		newp := randPointInRegion(r)
		ok := true
		for _, p := range r.points {
			if p.equal(newp) || dist(p, newp) <= minPlaceDist {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		fail = 0
		r.points = append(r.points, newp)
	}
	log.Printf("%d points generated\n", len(r.points))
}

func (u *Universe) generatePoints() {
	u.Region.generatePoints(u.MinPlaceDist)
}

func (r *Region) computeDists(maxWayLength float64) {
	for _, a := range r.points {
		for _, b := range r.points {
			if a.equal(b) {
				continue
			}
			if d := dist(a, b); d <= maxWayLength {
				r.dists[segment{a: a, b: b}] = d
				r.dists[segment{a: b, b: a}] = d
			}
		}
	}
	log.Printf("%d potential segments\n", len(r.dists))
}

func (u *Universe) computeDists() {
	u.Region.computeDists(u.MaxWayLength)
}

func copySegment(s segment) segment {
	a := point{x: s.a.x, y: s.a.y}
	b := point{x: s.b.x, y: s.b.y}

	return segment{a: a, b: b}
}

func (r *Region) generateSegments() {
	for _, a := range r.points {
		for news := range r.dists {
			if news.a.equal(a) {
				ok := true
				for _, s := range r.segments {
					if segmentIntersect(s, news) {
						ok = false
						break
					}
				}
				if !ok {
					continue
				}
				r.segments = append(r.segments, copySegment(news))
			}
		}
	}
	log.Printf("%d segments generated\n", len(r.segments))
}

func (u *Universe) generateSegments() {
	u.Region.generateSegments()
}

func (u *Universe) generateRegions() {
	u.Regions = make([]*Region, 0)

	for i := 0; i < u.RegionConfig.Count; i++ {
		for {
			p := u.Region.points[rand.Intn(len(u.Region.points))]
			ok := true
			for _, r := range u.Regions {
				if dist(r.Center, p) <= r.Radius {
					ok = false
					break
				}
			}
			if !ok {
				continue
			}
			u.Regions = append(u.Regions, newRegion(p, u.RegionConfig.Radius))
			break
		}
	}
	u.densifyRegions()
}

func (u *Universe) densifyRegions() {
	for _, r := range u.Regions {
		r.generatePoints(u.RegionConfig.MinPlaceDist)
		r.computeDists(u.RegionConfig.MaxWayLength)
		r.generateSegments()
	}
}

func (u *Universe) makePlacesAndWays() {
	np := len(u.Region.points)
	for _, r := range u.Regions {
		np += len(r.points)
	}
	u.Places = make([]Place, np)
	n := 0
	for _, p := range u.Region.points {
		u.Places[n] = newPlace(p)
		n++
	}
	for _, r := range u.Regions {
		for _, p := range r.points {
			u.Places[n] = newPlace(p)
			n++
		}
	}

	ns := len(u.Region.segments)
	for _, r := range u.Regions {
		ns += len(r.segments)
	}
	u.Ways = make([]Way, ns)
	n = 0
	for _, s := range u.Region.segments {
		u.Ways[n] = newWay(s, u.Places)
		n++
	}
	for _, r := range u.Regions {
		for _, s := range r.segments {
			u.Ways[n] = newWay(s, u.Places)
			n++
		}
	}
}

func (u *Universe) cleanup() {
	u.Region.points = make([]point, 0)
	u.Region.segments = make([]segment, 0)
	u.Region.dists = make(map[segment]float64)
	for _, r := range u.Regions {
		r.points = make([]point, 0)
		r.segments = make([]segment, 0)
		r.dists = make(map[segment]float64)
	}
}

// Generate generates a new random universe
func Generate(cfg Config) *Universe {
	u := newUniverse(cfg)

	log.Printf("Generating places at least distant from %d... ", int(u.MinPlaceDist))
	u.generatePoints()
	log.Printf("OK\n")

	log.Printf("Computing distances beneath %d... ", int(u.MaxWayLength))
	u.computeDists()
	log.Printf("OK\n")

	log.Printf("Computing ways... ")
	u.generateSegments()
	log.Printf("OK\n")

	log.Printf("Computing regions... ")
	u.generateRegions()
	log.Printf("OK\n")

	u.makePlacesAndWays()

	u.cleanup()

	return u
}
