package runtime

import "net/http"

type MatcherList []Matcher

func (ml MatcherList) Match(r http.Request) bool {
	for _, m := range ml {
		if !m.Match(r) {
			return false
		}
	}
	return true
}

type Matcher interface {
	Match(http.Request) bool
}

type PathMatcher struct {
	Path string
}

func (p PathMatcher) Match(r http.Request) bool {
	return r.URL.Path == p.Path
}
