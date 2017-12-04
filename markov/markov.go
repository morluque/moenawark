/*
Package markov implements a simple Markov Chain random name generator.

It is used to generate random names for universe places.
*/
package markov

import (
	"bufio"
	"github.com/morluque/moenawark/config"
	"github.com/morluque/moenawark/loglevel"
	"io"
	"math/rand"
	"unicode/utf8"
)

// Chains represents letter-based Markov chains, udes to generate random
// words.
type Chains struct {
	prefixLen int
	starts    []string
	digraphs  map[string]digraph
}

type digraph map[rune]float64

const endOfWord rune = 0

var log *loglevel.Logger

func init() {
	log = loglevel.New("markov", loglevel.Debug)
}

// ReloadConfig performs required actions to reload all dynamic config.
func ReloadConfig() {
	log.SetLevelName(config.Get("loglevel.markov"))
}

func newMarkovChains(prefixLen int) *Chains {
	starts := make([]string, 0)
	digraphs := make(map[string]digraph)
	return &Chains{prefixLen: prefixLen, digraphs: digraphs, starts: starts}
}

func (m *Chains) add(prefix string, suffix rune) {
	if d, ok := m.digraphs[prefix]; ok {
		d[suffix]++
	} else {
		m.digraphs[prefix] = make(map[rune]float64)
		m.digraphs[prefix][suffix] = 1
	}
}

func (m *Chains) normalize() {
	for _, d := range m.digraphs {
		var total float64
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

func analyzeWords(prefixLen int, words []string) *Chains {
	m := newMarkovChains(prefixLen)
	for n, w := range words {
		if !utf8.ValidString(w) {
			log.Fatalf("invalid UTF8 string %q at line %d", w, n+1)
			break
		}
		if len(w) <= m.prefixLen {
			continue
		}
		runes := getRunes(w)
		m.starts = append(m.starts, runesToString(runes[0:m.prefixLen]))
		var i int
		for i = m.prefixLen; i < len(runes)-1; i++ {
			if i >= len(runes) {
				break
			}
			m.add(runesToString(runes[i-m.prefixLen:i]), runes[i])
		}
		m.add(runesToString(runes[i-m.prefixLen:i]), endOfWord)
	}

	return m
}

// Load loads and analyzes a list of words (one per line) from a io.Reader.
// Once loaded, it is ready to generate random words.
// The prefixLen parameter is the number of characters that are associated with
// a probability of a following character.
func Load(r io.Reader, prefixLen int) *Chains {
	fscanner := bufio.NewScanner(r)
	words := make([]string, 0)
	for fscanner.Scan() {
		words = append(words, fscanner.Text())
	}
	m := analyzeWords(prefixLen, words)
	log.Debugf("loaded %d words", len(words))
	m.normalize()

	return m
}

func advancePrefix(prefix string, r rune) (rune, string) {
	prefixRunes := getRunes(prefix)
	prefixRunes = append(prefixRunes, r)
	newPrefixRunes := prefixRunes[1:]

	return prefixRunes[0], runesToString(newPrefixRunes)
}

// Generate returns a random word according to the probabilities stored in
// markov.Chains .
func (m *Chains) Generate() string {
	wordRunes := make([]rune, 0)
	prefix := m.starts[rand.Intn(len(m.starts))]
	var selectedRune rune // default value is 0
	for {
		p := rand.Float64()
		var n float64
		var ru rune // default value is 0
		for r, q := range m.digraphs[prefix] {
			n += q
			if p <= n {
				ru = r
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
