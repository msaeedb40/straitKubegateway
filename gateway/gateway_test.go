package gateway_test

import (
	"net/netip"
	"testing"

	"github.com/straitKubegateway/straitKubegateway/gateway"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

func TestHostnameMatching(t *testing.T) {
	tests := []struct {
		pattern string
		host    string
		match   bool
	}{
		{"example.com", "example.com", true},
		{"example.com", "other.com", false},
		{"*.example.com", "foo.example.com", true},
		{"*.example.com", "bar.example.com", true},
		{"*.example.com", "example.com", false},
		{"*", "anything.org", true},
		{"", "anything.org", true},
	}

	for _, tt := range tests {
		got := gateway.MatchesHostname(tt.pattern, tt.host)
		if got != tt.match {
			t.Errorf("MatchesHostname(%q, %q) = %v, expected %v", tt.pattern, tt.host, got, tt.match)
		}
	}
}

func TestHTTPRouteMatchingAndFilters(t *testing.T) {
	route := &gateway.Route{
		ID:        "default/api-route",
		Namespace: "default",
		Name:      "api-route",
		Hostnames: []string{"api.example.com"},
		Rules: []gateway.HTTPRouteRule{
			{
				Matches: []gateway.HTTPRouteMatch{
					{
						Path: &gateway.HTTPPathMatch{
							Type:  gateway.PathMatchPathPrefix,
							Value: "/v1/users",
						},
						Method: "GET",
						Headers: []gateway.HTTPHeaderMatch{
							{
								Name:  "X-Custom-Env",
								Value: "staging",
								Type:  gateway.HeaderMatchExact,
							},
						},
					},
				},
				Filters: []gateway.HTTPRouteFilter{
					{
						Type: "RequestHeaderModifier",
						SetHeaders: map[string]string{
							"X-Gateway-Enforced": "true",
						},
					},
				},
				Backends: []gateway.WeightedBackend{
					{
						IP:     netip.MustParseAddr("10.244.1.25"),
						Port:   8080,
						Weight: 100,
					},
				},
			},
		},
		Protocol: types.ProtocolTCP,
	}

	router := gateway.NewRouter([]*gateway.Route{route})

	// 1. Matching request
	ctxMatch := gateway.RequestContext{
		Hostname: "api.example.com",
		Path:     "/v1/users/12345",
		Method:   "GET",
		Headers: map[string]string{
			"X-Custom-Env": "staging",
		},
	}

	decision, err := router.RouteRequest(ctxMatch)
	if err != nil {
		t.Fatalf("expected match, got error: %v", err)
	}
	if decision.Backend.IP.String() != "10.244.1.25" || decision.Backend.Port != 8080 {
		t.Errorf("expected backend 10.244.1.25:8080, got %s:%d", decision.Backend.IP, decision.Backend.Port)
	}
	if decision.SetHeaders["X-Gateway-Enforced"] != "true" {
		t.Errorf("expected header filter X-Gateway-Enforced=true")
	}

	// 2. Non-matching header request -> Should fail
	ctxMismatch := gateway.RequestContext{
		Hostname: "api.example.com",
		Path:     "/v1/users/12345",
		Method:   "GET",
		Headers: map[string]string{
			"X-Custom-Env": "prod",
		},
	}
	_, errMismatch := router.RouteRequest(ctxMismatch)
	if errMismatch == nil {
		t.Errorf("expected routing failure on header mismatch")
	}
}

func TestWeightedTrafficSplitting(t *testing.T) {
	route := &gateway.Route{
		ID:        "default/canary-route",
		Namespace: "default",
		Name:      "canary-route",
		Hostnames: []string{"canary.example.com"},
		Rules: []gateway.HTTPRouteRule{
			{
				Backends: []gateway.WeightedBackend{
					{IP: netip.MustParseAddr("10.244.1.10"), Port: 8080, Weight: 80},
					{IP: netip.MustParseAddr("10.244.1.20"), Port: 8080, Weight: 20},
				},
			},
		},
		Protocol: types.ProtocolTCP,
	}

	router := gateway.NewRouter([]*gateway.Route{route})

	counts := make(map[string]int)
	total := 1000

	for i := 0; i < total; i++ {
		dec, err := router.RouteRequest(gateway.RequestContext{
			Hostname: "canary.example.com",
			Path:     "/",
		})
		if err != nil {
			t.Fatalf("RouteRequest failed: %v", err)
		}
		counts[dec.Backend.IP.String()]++
	}

	// 80/20 split expectation: ~800 to 10.244.1.10, ~200 to 10.244.1.20 (allow +/- 8% tolerance)
	v1Count := counts["10.244.1.10"]
	v2Count := counts["10.244.1.20"]

	if v1Count < 720 || v1Count > 880 {
		t.Errorf("v1 weight distribution out of tolerance (expected ~800, got %d)", v1Count)
	}
	if v2Count < 120 || v2Count > 280 {
		t.Errorf("v2 weight distribution out of tolerance (expected ~200, got %d)", v2Count)
	}
}
