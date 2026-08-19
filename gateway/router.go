package gateway

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
)

// RequestContext represents an inbound request evaluated by the router.
type RequestContext struct {
	Hostname string
	Path     string
	Method   string
	Headers  map[string]string
	Port     uint16
}

// RouteDecision represents the routing decision produced by the router.
type RouteDecision struct {
	Backend      WeightedBackend
	RedirectURL  string
	RewritePath  string
	SetHeaders   map[string]string
	AddHeaders   map[string]string
	MatchingRule *HTTPRouteRule
}

// Router executes route matching, header/path filters, and weighted traffic splitting.
type Router struct {
	mu     sync.RWMutex
	routes []*Route
}

// NewRouter creates a new Gateway API router.
func NewRouter(routes []*Route) *Router {
	return &Router{
		routes: routes,
	}
}

// RouteRequest evaluates an inbound request against configured routes and returns a RouteDecision.
func (r *Router) RouteRequest(ctx RequestContext) (*RouteDecision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		// 1. Match Hostname
		if !matchHostnames(route.Hostnames, ctx.Hostname) {
			continue
		}

		// 2. Evaluate Rules
		for _, rule := range route.Rules {
			if matchesRule(rule, ctx) {
				// Select weighted backend
				backend, err := selectWeightedBackend(rule.Backends)
				if err != nil && rule.RedirectURL == "" {
					return nil, err
				}

				decision := &RouteDecision{
					Backend:      backend,
					RedirectURL:  rule.RedirectURL,
					RewritePath:  rule.RewritePath,
					MatchingRule: &rule,
					SetHeaders:   make(map[string]string),
					AddHeaders:   make(map[string]string),
				}

				// Apply filters
				for _, f := range rule.Filters {
					for k, v := range f.SetHeaders {
						decision.SetHeaders[k] = v
					}
					for k, v := range f.AddHeaders {
						decision.AddHeaders[k] = v
					}
				}

				return decision, nil
			}
		}
	}

	return nil, fmt.Errorf("no matching route for host %q and path %q", ctx.Hostname, ctx.Path)
}

func matchHostnames(patterns []string, host string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if MatchesHostname(p, host) {
			return true
		}
	}
	return false
}

func matchesRule(rule HTTPRouteRule, ctx RequestContext) bool {
	if len(rule.Matches) == 0 {
		return true // Default rule
	}

	for _, m := range rule.Matches {
		if matchCriteria(m, ctx) {
			return true
		}
	}
	return false
}

func matchCriteria(m HTTPRouteMatch, ctx RequestContext) bool {
	// Method match
	if m.Method != "" && !strings.EqualFold(m.Method, ctx.Method) {
		return false
	}

	// Path match
	if m.Path != nil {
		switch m.Path.Type {
		case PathMatchExact:
			if ctx.Path != m.Path.Value {
				return false
			}
		case PathMatchPathPrefix:
			if !strings.HasPrefix(ctx.Path, m.Path.Value) {
				return false
			}
		case PathMatchRegularExpression:
			re, err := regexp.Compile(m.Path.Value)
			if err != nil || !re.MatchString(ctx.Path) {
				return false
			}
		}
	}

	// Headers match
	for _, hm := range m.Headers {
		val, exists := ctx.Headers[hm.Name]
		if !exists {
			return false
		}
		if hm.Type == HeaderMatchExact && val != hm.Value {
			return false
		}
		if hm.Type == HeaderMatchRegularExpression {
			re, err := regexp.Compile(hm.Value)
			if err != nil || !re.MatchString(val) {
				return false
			}
		}
	}

	return true
}

func selectWeightedBackend(backends []WeightedBackend) (WeightedBackend, error) {
	if len(backends) == 0 {
		return WeightedBackend{}, fmt.Errorf("no backends configured")
	}
	if len(backends) == 1 {
		return backends[0], nil
	}

	var totalWeight uint32
	for _, b := range backends {
		totalWeight += b.Weight
	}
	if totalWeight == 0 {
		return backends[0], nil
	}

	r := rand.Uint32() % totalWeight
	var acc uint32
	for _, b := range backends {
		acc += b.Weight
		if r < acc {
			return b, nil
		}
	}

	return backends[0], nil
}
