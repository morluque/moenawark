/*
Package universe implements random universe generation for Moenawark.
*/
package universe

import (
	"database/sql"
	"fmt"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"github.com/morluque/moenawark/markov"
	"github.com/morluque/moenawark/model"
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
	MarkovGen    *markov.Chains
}

// Universe stores places and ways between them
type Universe struct {
	Config
	Region    *Region
	Regions   []*Region
	Places    []*model.Place
	Wormholes []*model.Wormhole
	names     map[string]bool
}

var log *loglevel.Logger

func init() {
	log = loglevel.New("universe", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
	log.SetLevelName(config.Get("loglevel.universe"))
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
}

func newRegion(center point, radius float64) *Region {
	r := Region{
		Center: point{x: center.x, y: center.y},
		Radius: radius,
	}

	r.points = make([]point, 0)
	r.segments = make([]segment, 0)

	return &r
}

// WriteDotFile exports a universe to Graphviz "dot" format
func (u *Universe) WriteDotFile(path string) error {
	scale := 20
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
		_, err = f.WriteString(fmt.Sprintf("    p%d [pos=\"%d,%d!\"];\n", p.ID, p.X*scale, p.Y*scale))
		if err != nil {
			return err
		}
	}
	for _, w := range u.Wormholes {
		_, err = f.WriteString(fmt.Sprintf("    p%d -> p%d;\n", w.Source.ID, w.Destination.ID))
		if err != nil {
			return err
		}
	}
	_, err = f.WriteString("}\n")
	return err
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
	log.Infof("%d points generated\n", len(r.points))
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
	log.Infof("%d potential segments\n", len(dists))

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
	log.Infof("generated %d segments\n", len(segments))

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
		log.Infof("region [%f, %f] r%f\n", r.Center.x, r.Center.y, r.Radius)
		r.generatePoints(u.RegionConfig.MinPlaceDist, points)
		dists := computeDists(r.points, points, u.RegionConfig.MaxWayLength)
		r.segments = generateSegments(dists, segments)
	}
}

func placeFromPoint(p point, markovGen *markov.Chains, names map[string]bool) *model.Place {
	var name string
	for {
		name = markovGen.Generate()
		if _, found := names[name]; !found {
			break
		}
	}
	if len(name) == 0 {
		log.Fatal("markov random name is empty!")
	}
	names[name] = true
	return model.NewPlace(name, int(p.x), int(p.y))
}

func wormholeFromSegment(s segment, places []*model.Place) *model.Wormhole {
	srcFound, dstFound := false, false
	var src, dst *model.Place
	for _, p := range places {
		if p.X == int(s.a.x) && p.Y == int(s.a.y) {
			src = p
			srcFound = true
		}
		if p.X == int(s.b.x) && p.Y == int(s.b.y) {
			dst = p
			dstFound = true
		}
		if srcFound && dstFound {
			break
		}
	}

	return model.NewWormhole(src, dst, int(dist(s.a, s.b)))
}

func (u *Universe) makePlacesAndWormholes(tx *sql.Tx) error {
	u.names = make(map[string]bool)

	np := len(u.Region.points)
	for _, r := range u.Regions {
		np += len(r.points)
	}
	u.Places = make([]*model.Place, np)
	n := 0
	for _, p := range u.Region.points {
		u.Places[n] = placeFromPoint(p, u.MarkovGen, u.names)
		n++
	}
	for _, r := range u.Regions {
		for _, p := range r.points {
			u.Places[n] = placeFromPoint(p, u.MarkovGen, u.names)
			n++
		}
	}
	for _, p := range u.Places {
		if err := p.Save(tx); err != nil {
			return err
		}
	}
	log.Infof("Saved %d places to database\n", len(u.Places))

	ns := len(u.Region.segments)
	for _, r := range u.Regions {
		ns += len(r.segments)
	}
	u.Wormholes = make([]*model.Wormhole, ns)
	n = 0
	for _, s := range u.Region.segments {
		u.Wormholes[n] = wormholeFromSegment(s, u.Places)
		n++
	}
	for _, r := range u.Regions {
		for _, s := range r.segments {
			u.Wormholes[n] = wormholeFromSegment(s, u.Places)
			n++
		}
	}
	for _, w := range u.Wormholes {
		if err := w.Save(tx); err != nil {
			return err
		}
	}

	return nil
}

func (u *Universe) cleanup() {
	u.Region.points = make([]point, 0)
	u.Region.segments = make([]segment, 0)
	for _, r := range u.Regions {
		r.points = make([]point, 0)
		r.segments = make([]segment, 0)
	}
}

// Generate generates a new random universe
func Generate(cfg Config, tx *sql.Tx) *Universe {
	u := newUniverse(cfg)

	log.Infof("Computing regions...")
	u.generateRegions()

	log.Infof("Generating places at least distant from %d...", int(u.MinPlaceDist))
	u.generatePoints()

	log.Infof("Computing ways...")
	u.generateSegments()

	if err := u.makePlacesAndWormholes(tx); err != nil {
		log.Fatal(err)
	}

	u.cleanup()

	return u
}
