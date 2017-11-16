package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"unicode/utf8"
)

type MarkovChains struct {
	prefixes []string
	digraphs map[string]digraph
}

type digraph map[rune]float64

const EOW rune = 0

func newMarkovChains() *MarkovChains {
	prefixes := make([]string, 0)
	digraphs := make(map[string]digraph)
	return &MarkovChains{digraphs: digraphs}
}

func (m *MarkovChains) Add(prefix string, suffix rune) {
	if d, ok := m.digraphs[prefix]; ok {
		d[suffix] += 1
	} else {
		m.digraphs[prefix] = make(map[rune]float64)
		m.digraphs[prefix][suffix] = 1
	}
}

func (m *MarkovChains) Normalize() {
	for prefix, d := range m.digraphs {
		m.prefixes = append(m.prefixes, prefix)
		var total float64 = 0
		runes := make([]rune, 0)
		for r, count := range d {
			total += count
			runes = append(runes, r)
		}
		for _, r := range runes {
			d[r] = d[r] / total
		}
	}
}

func getRunes(s string) []rune {
	runes := make([]rune, utf8.RuneCountInString(s))
	i := 0
	for _, r := range s {
		runes[i] = r
		i++
	}

	return runes
}

func runesToString(runes []rune) string {
	bytes := make([]byte, 0)
	for _, r := range runes {
		n := utf8.RuneLen(r)
		b := make([]byte, n)
		utf8.EncodeRune(b, r)
		bytes = append(bytes, b...)
	}

	return string(bytes)
}

func Analyze(words []string) *MarkovChains {
	m := newMarkovChains()
	for n, w := range words {
		if !utf8.ValidString(w) {
			log.Fatal("invalid UTF8 string %q at line %d", w, n+1)
			break
		}
		if len(w) < 3 {
			continue
		}
		runes := getRunes(w)
		var i int
		for i = 2; i < len(runes)-1; i++ {
			if i >= len(runes) {
				break
			}
			m.Add(runesToString(runes[i-2:i]), runes[i])
		}
		m.Add(runesToString(runes[i-2:i]), EOW)
	}

	return m
}

func (m *MarkovChains) Generate() string {
	prefix := m.prefixes[rand.Intn(len(m.prefixes))]
	for {
		p := rand.Float64()
		var n float64 = 0
		for r, q := range m.digraphs[prefix] {
			n += q
			if p <= q {
			}
		}
	}
	return ""
}

func main() {
	fscanner := bufio.NewScanner(os.Stdin)
	words := make([]string, 0)
	for fscanner.Scan() {
		words = append(words, fscanner.Text())
	}
	m := Analyze(words)
	m.Normalize()
	fmt.Printf("generated: %s\n", m.Generate())
}
