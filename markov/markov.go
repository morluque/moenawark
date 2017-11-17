package markov

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"unicode/utf8"
)

type MarkovChains struct {
	starts   []string
	digraphs map[string]digraph
}

type digraph map[rune]float64

const EOW rune = 0
const PREFIXLEN = 2

func newMarkovChains() *MarkovChains {
	starts := make([]string, 0)
	digraphs := make(map[string]digraph)
	return &MarkovChains{digraphs: digraphs, starts: starts}
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
	for _, d := range m.digraphs {
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
		if len(w) <= PREFIXLEN {
			continue
		}
		runes := getRunes(w)
		m.starts = append(m.starts, runesToString(runes[0:PREFIXLEN]))
		var i int
		for i = PREFIXLEN; i < len(runes)-1; i++ {
			if i >= len(runes) {
				break
			}
			m.Add(runesToString(runes[i-PREFIXLEN:i]), runes[i])
		}
		m.Add(runesToString(runes[i-PREFIXLEN:i]), EOW)
	}

	return m
}

func advancePrefix(prefix string, r rune) (rune, string) {
	prefixRunes := getRunes(prefix)
	prefixRunes = append(prefixRunes, r)
	newPrefixRunes := prefixRunes[1:]

	return prefixRunes[0], runesToString(newPrefixRunes)
}

func (m *MarkovChains) Generate() string {
	wordRunes := make([]rune, 0)
	prefix := m.starts[rand.Intn(len(m.starts))]
	var selectedRune rune = 0
	for {
		p := rand.Float64()
		var n float64 = 0
		var ru rune = 0
		for r, q := range m.digraphs[prefix] {
			n += q
			if p <= n {
				ru = r
				//log.Printf("%#U, p=%f", ru, q)
				break
			}
		}
		if ru == 0 {
			// end of word
			wordRunes = append(wordRunes, getRunes(prefix)...)
			break
		} else {
			selectedRune, prefix = advancePrefix(prefix, ru)
			wordRunes = append(wordRunes, selectedRune)
		}
	}
	return runesToString(wordRunes)
}

/*
func main() {
	fscanner := bufio.NewScanner(os.Stdin)
	words := make([]string, 0)
	for fscanner.Scan() {
		words = append(words, fscanner.Text())
	}
	m := Analyze(words)
	m.Normalize()
	fmt.Printf("generated: %s\n", m.Generate())
	fmt.Printf("generated: %s\n", m.Generate())
}
*/
