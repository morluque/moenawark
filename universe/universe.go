package universe

import (
	"fmt"
	"log"
	"math/rand"
	"os"
)

// RegionConfig holds configuration for a universe region
type RegionConfig struct {
	Count        int
	Radius       float64
	MinPlaceDist float64
	MaxWayLength float64
}

// Config holds configuration for a random universe
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

// Region represent a circular region of universe with more places density
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

func (r *Region) containsPoint(p point) bool {
	return dist(r.Center, p) <= r.Radius
}

func randPointInRegion(r *Region) point {
	for i := 0; i < 1000; i++ {
		x := r.Center.x - r.Radius + rand.Float64()*r.Radius*2
		y := r.Center.y - r.Radius + rand.Float64()*r.Radius*2
		p := point{x: x, y: y}
		if r.containsPoint(p) {
			return p
		}
	}
	return point{}
}

func (r *Region) generatePoints(minPlaceDist float64, otherPoints []point) {
	fail := 0
	for {
		fail++
		if fail > int(r.Radius)*100 {
			break
		}
		newp := randPointInRegion(r)
		if newp.farEnough(minPlaceDist, otherPoints...) && newp.farEnough(minPlaceDist, r.points...) {
			fail = 0
			r.points = append(r.points, newp)
		}
	}
	log.Printf("%d points generated\n", len(r.points))
}

func (u *Universe) generatePoints() {
	points := make([]point, 0)
	for _, r := range u.Regions {
		for _, p := range r.points {
			points = append(points, p)
		}
	}
	u.Region.generatePoints(u.MinPlaceDist, points)
}

func computeDists(srcs, dsts []point, maxWayLength float64) map[segment]float64 {
	dists := make(map[segment]float64)
	srcdsts := append(srcs, dsts...)
	for _, a := range srcs {
		for _, b := range srcdsts {
			if a.equal(b) {
				continue
			}
			if d := dist(a, b); d <= maxWayLength {
				dists[segment{a: a, b: b}] = d
			}
		}
	}
	log.Printf("%d potential segments\n", len(dists))

	return dists
}

func generateSegments(dists map[segment]float64, existingSegments []segment) []segment {
	segments := make([]segment, 0)

	for news := range dists {
		if news.intersect(existingSegments...) || news.intersect(segments...) {
			continue
		}
		segments = append(segments, news)
	}
	log.Printf("generated %d segments\n", len(segments))

	return segments
}

func (u *Universe) generateSegments() {
	sources := u.Region.points
	dests := make([]point, 0)
	existingSegments := make([]segment, 0)
	for _, r := range u.Regions {
		dests = append(dests, r.points...)
		existingSegments = append(existingSegments, r.segments...)
	}
	dists := computeDists(sources, dests, u.MaxWayLength)
	u.Region.segments = generateSegments(dists, existingSegments)
}

func (u *Universe) generateRegions() {
	u.Regions = make([]*Region, 0)

	for i := 0; i < u.RegionConfig.Count; i++ {
		for {
			p := randPointInRegion(u.Region)
			ok := true
			for _, r := range u.Regions {
				if dist(r.Center, p) <= r.Radius*2 {
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
	points := make([]point, 0)
	segments := make([]segment, 0)
	for _, r := range u.Regions {
		log.Printf("region [%f, %f] r%f\n", r.Center.x, r.Center.y, r.Radius)
		r.generatePoints(u.RegionConfig.MinPlaceDist, points)
		dists := computeDists(r.points, points, u.RegionConfig.MaxWayLength)
		r.segments = generateSegments(dists, segments)
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

	log.Printf("Computing regions... ")
	u.generateRegions()
	log.Printf("OK\n")

	log.Printf("Generating places at least distant from %d... ", int(u.MinPlaceDist))
	u.generatePoints()
	log.Printf("OK\n")

	log.Printf("Computing ways... ")
	u.generateSegments()
	log.Printf("OK\n")

	u.makePlacesAndWays()

	u.cleanup()

	return u
}
