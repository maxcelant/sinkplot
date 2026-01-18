package routes

import (
	"net/http"
	"regexp"
	"strings"
)

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

// PathMatcher matches exact paths
type PathMatcher struct {
	Path string
}

func (p PathMatcher) Match(r http.Request) bool {
	return r.URL.Path == p.Path
}

// PrefixMatcher matches paths that start with the given prefix
type PrefixMatcher struct {
	Prefix string
}

func (p PrefixMatcher) Match(r http.Request) bool {
	return strings.HasPrefix(r.URL.Path, p.Prefix)
}

// RegexMatcher matches paths against a regular expression
type RegexMatcher struct {
	Pattern *regexp.Regexp
}

func (r RegexMatcher) Match(req http.Request) bool {
	return r.Pattern.MatchString(req.URL.Path)
}

// MethodMatcher matches HTTP methods
type MethodMatcher struct {
	Methods []string
}

func (m MethodMatcher) Match(r http.Request) bool {
	for _, method := range m.Methods {
		if strings.EqualFold(r.Method, method) {
			return true
		}
	}
	return false
}
