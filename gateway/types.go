// Package gateway provides Gateway API routing, HTTPRoute/GRPCRoute/TCPRoute/UDPRoute
// matching engines, filters, and traffic splitting for straitKubegateway.
package gateway

import (
	"net/netip"
	"strings"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// PathMatchType defines how the HTTP path is matched.
type PathMatchType string

const (
	PathMatchExact             PathMatchType = "Exact"
	PathMatchPathPrefix        PathMatchType = "PathPrefix"
	PathMatchRegularExpression PathMatchType = "RegularExpression"
)

// HeaderMatchType defines how an HTTP header is matched.
type HeaderMatchType string

const (
	HeaderMatchExact             HeaderMatchType = "Exact"
	HeaderMatchRegularExpression HeaderMatchType = "RegularExpression"
)

// HTTPHeaderMatch defines a header match criteria.
type HTTPHeaderMatch struct {
	Name  string
	Value string
	Type  HeaderMatchType
}

// HTTPPathMatch defines a path match criteria.
type HTTPPathMatch struct {
	Type  PathMatchType
	Value string
}

// HTTPRouteRule defines a single routing rule within an HTTPRoute.
type HTTPRouteRule struct {
	Matches     []HTTPRouteMatch
	Filters     []HTTPRouteFilter
	Backends    []WeightedBackend
	RedirectURL string
	RewritePath string
}

// HTTPRouteMatch defines criteria to match an HTTP request.
type HTTPRouteMatch struct {
	Path    *HTTPPathMatch
	Headers []HTTPHeaderMatch
	Method  string
}

// HTTPRouteFilter defines request modification filters.
type HTTPRouteFilter struct {
	Type          string // RequestHeaderModifier, ResponseHeaderModifier, URLRewrite, RequestRedirect, RequestMirror
	SetHeaders    map[string]string
	AddHeaders    map[string]string
	RemoveHeaders []string
}

// WeightedBackend defines a destination backend with traffic weight.
type WeightedBackend struct {
	IP     netip.Addr
	Port   uint16
	Weight uint32
}

// Route represents a generic Gateway API route.
type Route struct {
	ID        string
	Namespace string
	Name      string
	Hostnames []string
	Rules     []HTTPRouteRule
	Protocol  types.Protocol
}

// GatewayInstance represents an active Gateway resource with its listeners.
type GatewayInstance struct {
	ID        string
	Namespace string
	Name      string
	SegmentID types.SegmentID
	Listeners []Listener
	Routes    map[string]*Route // "namespace/name" -> Route
}

// Listener represents a port/protocol listener on a Gateway.
type Listener struct {
	Name     string
	Port     uint16
	Protocol string // HTTP, HTTPS, TLS, TCP, UDP, gRPC
	Hostname string
	TLS      *TLSConfig
}

// TLSConfig defines TLS termination/passthrough configuration.
type TLSConfig struct {
	Mode           string // Terminate, Passthrough
	CertificateRef string
}

// MatchesHostname checks if a request hostname matches a listener or route hostname (supports wildcards like *.example.com).
func MatchesHostname(pattern, host string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	pattern = strings.ToLower(pattern)
	host = strings.ToLower(host)

	if pattern == host {
		return true
	}

	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // e.g. ".example.com"
		return strings.HasSuffix(host, suffix)
	}

	return false
}
