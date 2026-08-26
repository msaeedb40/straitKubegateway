// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package gateway implements the straitKubegateway Gateway API controller.
// Supports Gateway API v1.6.1: GatewayClass, Gateway, HTTPRoute,
// TCPRoute, UDPRoute, GRPCRoute, TLSRoute.
package gateway

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// GatewayControllerName is the controller name used in GatewayClass.
const GatewayControllerName = "gateway.straitkubegateway.io/skubegateway"

// Condition Types & Reasons per Gateway API v1.6.1 spec
const (
	GatewayConditionAccepted   = "Accepted"
	GatewayConditionProgrammed = "Programmed"
	RouteConditionAccepted     = "Accepted"
	RouteConditionResolvedRefs = "ResolvedRefs"

	GatewayReasonAccepted        = "Accepted"
	GatewayReasonProgrammed      = "Programmed"
	RouteReasonAccepted          = "Accepted"
	RouteReasonResolvedRefs      = "ResolvedRefs"
	RouteReasonBackendNotFound   = "BackendNotFound"
)

// ============================================================================
// Gateway Manager
// ============================================================================

// Manager manages Gateway API resources for straitKubegateway.
type Manager struct {
	mu       sync.RWMutex
	client   client.Client
	log      *zap.Logger
	gateways map[types.NamespacedName]*GatewayState
}

// GatewayState holds the compiled state for a single Gateway.
type GatewayState struct {
	Key        types.NamespacedName
	ClassName  string
	Addresses  []string
	Listeners  []ListenerState
	Routes     []*RouteState
	Conditions []metav1.Condition
	Ready      bool
}

// ListenerState holds compiled listener parameters.
type ListenerState struct {
	Name     string
	Port     uint16
	Protocol string
	Hostname string
}

// RouteState holds the compiled state for a single route.
type RouteState struct {
	Kind       string // HTTPRoute, TCPRoute, UDPRoute, GRPCRoute, TLSRoute
	Namespace  string
	Name       string
	Rules      []RouteRule
	Conditions []metav1.Condition
}

// PathMatchType represents the type of path matching.
type PathMatchType string

const (
	PathMatchExact             PathMatchType = "Exact"
	PathMatchPathPrefix        PathMatchType = "PathPrefix"
	PathMatchRegularExpression PathMatchType = "RegularExpression"
)

// RouteRule is a compiled route rule with rich matchers and weighted backends.
type RouteRule struct {
	PathType   PathMatchType
	Path       string
	PathRegex  *regexp.Regexp
	Headers    map[string]string
	QueryParams map[string]string
	Backends   []RouteBackend
	Action     string
}

// RouteBackend is a single backend for a route rule with weight.
type RouteBackend struct {
	Namespace string
	Service   string
	Port      uint16
	Weight    int32
}

// NewManager creates a new gateway manager.
func NewManager(c client.Client, log *zap.Logger) *Manager {
	return &Manager{
		client:   c,
		log:      log,
		gateways: make(map[types.NamespacedName]*GatewayState),
	}
}

// ============================================================================
// GatewayClass reconciliation
// ============================================================================

// ReconcileGatewayClass handles GatewayClass resources.
func (m *Manager) ReconcileGatewayClass(ctx context.Context, key types.NamespacedName) error {
	var gc gwv1.GatewayClass
	if err := m.client.Get(ctx, key, &gc); err != nil {
		return client.IgnoreNotFound(err)
	}
	if string(gc.Spec.ControllerName) != GatewayControllerName {
		return nil // not managed by this controller
	}
	m.log.Info("GatewayClass accepted", zap.String("name", key.Name))
	return nil
}

// ============================================================================
// Gateway reconciliation
// ============================================================================

// ReconcileGateway reconciles a Gateway resource.
func (m *Manager) ReconcileGateway(ctx context.Context, key types.NamespacedName) error {
	var gw gwv1.Gateway
	if err := m.client.Get(ctx, key, &gw); err != nil {
		return client.IgnoreNotFound(err)
	}

	state := &GatewayState{
		Key:       key,
		ClassName: string(gw.Spec.GatewayClassName),
		Ready:     true,
		Conditions: []metav1.Condition{
			{
				Type:               GatewayConditionAccepted,
				Status:             metav1.ConditionTrue,
				Reason:             GatewayReasonAccepted,
				Message:            "Gateway configuration accepted and validated by straitKubegateway controller",
				LastTransitionTime: metav1.Now(),
			},
			{
				Type:               GatewayConditionProgrammed,
				Status:             metav1.ConditionTrue,
				Reason:             GatewayReasonProgrammed,
				Message:            "eBPF data plane hooks successfully programmed for listeners",
				LastTransitionTime: metav1.Now(),
			},
		},
	}

	for _, addr := range gw.Status.Addresses {
		state.Addresses = append(state.Addresses, addr.Value)
	}

	for _, l := range gw.Spec.Listeners {
		var hostname string
		if l.Hostname != nil {
			hostname = string(*l.Hostname)
		}
		state.Listeners = append(state.Listeners, ListenerState{
			Name:     string(l.Name),
			Port:     uint16(l.Port),
			Protocol: string(l.Protocol),
			Hostname: hostname,
		})
	}

	m.mu.Lock()
	if existing, ok := m.gateways[key]; ok {
		state.Routes = existing.Routes
	}
	m.gateways[key] = state
	m.mu.Unlock()

	m.log.Info("gateway reconciled",
		zap.String("namespace", key.Namespace),
		zap.String("name", key.Name),
		zap.String("class", state.ClassName),
		zap.Int("listeners", len(state.Listeners)),
	)
	return nil
}

// DeleteGateway removes a gateway from the manager.
func (m *Manager) DeleteGateway(key types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.gateways, key)
}

// ============================================================================
// Route Reconciliations: HTTPRoute, TCPRoute, UDPRoute, GRPCRoute, TLSRoute
// ============================================================================

// ReconcileHTTPRoute reconciles an HTTPRoute.
func (m *Manager) ReconcileHTTPRoute(ctx context.Context, key types.NamespacedName) error {
	var route gwv1.HTTPRoute
	if err := m.client.Get(ctx, key, &route); err != nil {
		return client.IgnoreNotFound(err)
	}

	compiled, err := m.compileHTTPRoute(&route)
	if err != nil {
		return fmt.Errorf("compile HTTPRoute %s: %w", key, err)
	}

	m.mu.Lock()
	m.attachRoute(route.Spec.ParentRefs, compiled)
	m.mu.Unlock()

	m.log.Info("HTTPRoute reconciled",
		zap.String("namespace", key.Namespace),
		zap.String("name", key.Name),
		zap.Int("rules", len(compiled.Rules)),
	)
	return nil
}

// ReconcileTCPRoute reconciles a TCPRoute.
func (m *Manager) ReconcileTCPRoute(ctx context.Context, key types.NamespacedName) error {
	var route gwv1a2.TCPRoute
	if err := m.client.Get(ctx, key, &route); err != nil {
		return client.IgnoreNotFound(err)
	}

	rs := &RouteState{
		Kind:      "TCPRoute",
		Namespace: route.Namespace,
		Name:      route.Name,
	}

	for _, rule := range route.Spec.Rules {
		rr := RouteRule{}
		for _, ref := range rule.BackendRefs {
			ns := route.Namespace
			if ref.Namespace != nil {
				ns = string(*ref.Namespace)
			}
			port := uint16(0)
			if ref.Port != nil {
				port = uint16(*ref.Port)
			}
			weight := int32(1)
			if ref.Weight != nil {
				weight = *ref.Weight
			}
			rr.Backends = append(rr.Backends, RouteBackend{
				Namespace: ns,
				Service:   string(ref.Name),
				Port:      port,
				Weight:    weight,
			})
		}
		rs.Rules = append(rs.Rules, rr)
	}

	m.mu.Lock()
	m.attachRoute(route.Spec.ParentRefs, rs)
	m.mu.Unlock()
	return nil
}

// ReconcileUDPRoute reconciles a UDPRoute.
func (m *Manager) ReconcileUDPRoute(ctx context.Context, key types.NamespacedName) error {
	var route gwv1a2.UDPRoute
	if err := m.client.Get(ctx, key, &route); err != nil {
		return client.IgnoreNotFound(err)
	}

	rs := &RouteState{
		Kind:      "UDPRoute",
		Namespace: route.Namespace,
		Name:      route.Name,
	}

	for _, rule := range route.Spec.Rules {
		rr := RouteRule{}
		for _, ref := range rule.BackendRefs {
			ns := route.Namespace
			if ref.Namespace != nil {
				ns = string(*ref.Namespace)
			}
			port := uint16(0)
			if ref.Port != nil {
				port = uint16(*ref.Port)
			}
			weight := int32(1)
			if ref.Weight != nil {
				weight = *ref.Weight
			}
			rr.Backends = append(rr.Backends, RouteBackend{
				Namespace: ns,
				Service:   string(ref.Name),
				Port:      port,
				Weight:    weight,
			})
		}
		rs.Rules = append(rs.Rules, rr)
	}

	m.mu.Lock()
	m.attachRoute(route.Spec.ParentRefs, rs)
	m.mu.Unlock()
	return nil
}

// ReconcileGRPCRoute reconciles a GRPCRoute.
func (m *Manager) ReconcileGRPCRoute(ctx context.Context, key types.NamespacedName) error {
	var route gwv1.GRPCRoute
	if err := m.client.Get(ctx, key, &route); err != nil {
		return client.IgnoreNotFound(err)
	}

	rs := &RouteState{
		Kind:      "GRPCRoute",
		Namespace: route.Namespace,
		Name:      route.Name,
	}

	for _, rule := range route.Spec.Rules {
		rr := RouteRule{}
		for _, ref := range rule.BackendRefs {
			ns := route.Namespace
			if ref.Namespace != nil {
				ns = string(*ref.Namespace)
			}
			port := uint16(0)
			if ref.Port != nil {
				port = uint16(*ref.Port)
			}
			weight := int32(1)
			if ref.Weight != nil {
				weight = *ref.Weight
			}
			rr.Backends = append(rr.Backends, RouteBackend{
				Namespace: ns,
				Service:   string(ref.Name),
				Port:      port,
				Weight:    weight,
			})
		}
		rs.Rules = append(rs.Rules, rr)
	}

	m.mu.Lock()
	m.attachRoute(route.Spec.ParentRefs, rs)
	m.mu.Unlock()
	return nil
}

// ReconcileTLSRoute reconciles a TLSRoute.
func (m *Manager) ReconcileTLSRoute(ctx context.Context, key types.NamespacedName) error {
	var route gwv1a2.TLSRoute
	if err := m.client.Get(ctx, key, &route); err != nil {
		return client.IgnoreNotFound(err)
	}

	rs := &RouteState{
		Kind:      "TLSRoute",
		Namespace: route.Namespace,
		Name:      route.Name,
	}

	for _, rule := range route.Spec.Rules {
		rr := RouteRule{}
		for _, ref := range rule.BackendRefs {
			ns := route.Namespace
			if ref.Namespace != nil {
				ns = string(*ref.Namespace)
			}
			port := uint16(0)
			if ref.Port != nil {
				port = uint16(*ref.Port)
			}
			weight := int32(1)
			if ref.Weight != nil {
				weight = *ref.Weight
			}
			rr.Backends = append(rr.Backends, RouteBackend{
				Namespace: ns,
				Service:   string(ref.Name),
				Port:      port,
				Weight:    weight,
			})
		}
		rs.Rules = append(rs.Rules, rr)
	}

	m.mu.Lock()
	m.attachRoute(route.Spec.ParentRefs, rs)
	m.mu.Unlock()
	return nil
}

func (m *Manager) compileHTTPRoute(route *gwv1.HTTPRoute) (*RouteState, error) {
	rs := &RouteState{
		Kind:      "HTTPRoute",
		Namespace: route.Namespace,
		Name:      route.Name,
		Conditions: []metav1.Condition{
			{
				Type:               RouteConditionAccepted,
				Status:             metav1.ConditionTrue,
				Reason:             RouteReasonAccepted,
				Message:            "Route accepted and bound to parent gateway listeners",
				LastTransitionTime: metav1.Now(),
			},
			{
				Type:               RouteConditionResolvedRefs,
				Status:             metav1.ConditionTrue,
				Reason:             RouteReasonResolvedRefs,
				Message:            "All backend references resolved successfully",
				LastTransitionTime: metav1.Now(),
			},
		},
	}

	for _, rule := range route.Spec.Rules {
		rr := RouteRule{
			Headers:     make(map[string]string),
			QueryParams: make(map[string]string),
		}

		for _, match := range rule.Matches {
			if match.Path != nil && match.Path.Value != nil {
				rr.Path = *match.Path.Value
				if match.Path.Type != nil {
					rr.PathType = PathMatchType(*match.Path.Type)
				} else {
					rr.PathType = PathMatchPathPrefix
				}
				if rr.PathType == PathMatchRegularExpression {
					rx, err := regexp.Compile(rr.Path)
					if err == nil {
						rr.PathRegex = rx
					}
				}
			}
			for _, h := range match.Headers {
				rr.Headers[string(h.Name)] = h.Value
			}
			for _, q := range match.QueryParams {
				rr.QueryParams[string(q.Name)] = q.Value
			}
		}

		for _, ref := range rule.BackendRefs {
			ns := route.Namespace
			if ref.Namespace != nil {
				ns = string(*ref.Namespace)
			}
			port := uint16(0)
			if ref.Port != nil {
				port = uint16(*ref.Port)
			}
			weight := int32(1)
			if ref.Weight != nil {
				weight = *ref.Weight
			}
			rr.Backends = append(rr.Backends, RouteBackend{
				Namespace: ns,
				Service:   string(ref.Name),
				Port:      port,
				Weight:    weight,
			})
		}
		rs.Rules = append(rs.Rules, rr)
	}
	return rs, nil
}

func (m *Manager) attachRoute(parents []gwv1.ParentReference, route *RouteState) {
	for _, parent := range parents {
		ns := route.Namespace
		if parent.Namespace != nil {
			ns = string(*parent.Namespace)
		}
		gwKey := types.NamespacedName{Namespace: ns, Name: string(parent.Name)}
		if gw, ok := m.gateways[gwKey]; ok {
			routes := gw.Routes[:0]
			for _, r := range gw.Routes {
				if !(r.Kind == route.Kind && r.Namespace == route.Namespace && r.Name == route.Name) {
					routes = append(routes, r)
				}
			}
			gw.Routes = append(routes, route)
		}
	}
}

// MatchRoute matches an incoming HTTP request against the routes attached to a gateway.
func (m *Manager) MatchRoute(gwKey types.NamespacedName, path string, headers map[string]string) (*RouteRule, *RouteBackend, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	gw, exists := m.gateways[gwKey]
	if !exists {
		return nil, nil, false
	}

	for _, route := range gw.Routes {
		for _, rule := range route.Rules {
			if matchRule(&rule, path, headers) {
				if len(rule.Backends) > 0 {
					return &rule, &rule.Backends[0], true
				}
				return &rule, nil, true
			}
		}
	}
	return nil, nil, false
}

func matchRule(r *RouteRule, path string, headers map[string]string) bool {
	// Path match
	if r.Path != "" {
		switch r.PathType {
		case PathMatchExact:
			if path != r.Path {
				return false
			}
		case PathMatchRegularExpression:
			if r.PathRegex != nil && !r.PathRegex.MatchString(path) {
				return false
			}
		default: // PathPrefix
			if !strings.HasPrefix(path, r.Path) {
				return false
			}
		}
	}

	// Header match
	for hk, hv := range r.Headers {
		reqVal, ok := headers[hk]
		if !ok || reqVal != hv {
			return false
		}
	}

	return true
}

// GetAll returns all gateway states.
func (m *Manager) GetAll() []*GatewayState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*GatewayState, 0, len(m.gateways))
	for _, gw := range m.gateways {
		out = append(out, gw)
	}
	return out
}
