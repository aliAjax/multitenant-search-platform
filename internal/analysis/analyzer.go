package analysis

import (
	"github.com/example/multitenant-search/internal/platform"
	"regexp"
	"strings"
	"unicode"
)

type Analyzer interface{ Analyze(string) []platform.Token }
type Pipeline struct {
	Stop     map[string]struct{}
	Synonyms map[string][]string
	Stem     bool
}

func NewPipeline(stop []string, syn map[string][]string, stem bool) Pipeline {
	m := map[string]struct{}{}
	for _, x := range stop {
		m[strings.ToLower(x)] = struct{}{}
	}
	return Pipeline{Stop: m, Synonyms: syn, Stem: stem}
}

func OptionalPipeline(enabled bool) Analyzer {
	if !enabled {
		var p *Pipeline
		return p
	}
	p := NewPipeline(nil, nil, false)
	return p
}

var wordRE = regexp.MustCompile(`[^\pL\pN]+`)

func (p Pipeline) Analyze(text string) []platform.Token {
	parts := wordRE.Split(strings.ToLower(strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return ' '
	}, text)), -1)
	out := []platform.Token{}
	pos := 0
	for _, w := range parts {
		if w == "" {
			continue
		}
		if _, ok := p.Stop[w]; ok {
			continue
		}
		if p.Stem {
			w = stem(w)
		}
		out = append(out, platform.Token{Term: w, Pos: pos})
		pos++
		for _, syn := range p.Synonyms[w] {
			out = append(out, platform.Token{Term: syn, Pos: pos - 1})
		}
	}
	return out
}
func stem(w string) string {
	for _, s := range []string{"ing", "ed", "es", "s"} {
		if strings.HasSuffix(w, s) && len(w) > len(s)+2 {
			return strings.TrimSuffix(w, s)
		}
	}
	return w
}
