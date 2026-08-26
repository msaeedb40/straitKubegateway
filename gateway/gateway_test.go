// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"testing"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayAPIRouteMatching(t *testing.T) {
	log := zap.NewNop()
	scheme := runtime.NewScheme()
	_ = gwv1.Install(scheme)

	gwName := "strait-gw"
	ns := "default"
	pathVal := "/api/v1/orders"
	prefixType := gwv1.PathMatchPathPrefix
	port := gwv1.PortNumber(8080)

	gwObj := &gwv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      gwName,
		},
		Spec: gwv1.GatewaySpec{
			GatewayClassName: "strait-class",
		},
	}

	routeObj := &gwv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      "orders-route",
		},
		Spec: gwv1.HTTPRouteSpec{
			CommonRouteSpec: gwv1.CommonRouteSpec{
				ParentRefs: []gwv1.ParentReference{
					{
						Name: gwv1.ObjectName(gwName),
					},
				},
			},
			Rules: []gwv1.HTTPRouteRule{
				{
					Matches: []gwv1.HTTPRouteMatch{
						{
							Path: &gwv1.HTTPPathMatch{
								Type:  &prefixType,
								Value: &pathVal,
							},
						},
					},
					BackendRefs: []gwv1.HTTPBackendRef{
						{
							BackendRef: gwv1.BackendRef{
								BackendObjectReference: gwv1.BackendObjectReference{
									Name: "orders-svc",
									Port: &port,
								},
							},
						},
					},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gwObj, routeObj).Build()
	mgr := NewManager(client, log)
	ctx := context.Background()

	gwKey := types.NamespacedName{Namespace: ns, Name: gwName}
	if err := mgr.ReconcileGateway(ctx, gwKey); err != nil {
		t.Fatalf("reconcile gateway failed: %v", err)
	}

	routeKey := types.NamespacedName{Namespace: ns, Name: "orders-route"}
	if err := mgr.ReconcileHTTPRoute(ctx, routeKey); err != nil {
		t.Fatalf("reconcile route failed: %v", err)
	}

	// Match positive path
	_, be, matched := mgr.MatchRoute(gwKey, "/api/v1/orders/1234", nil)
	if !matched || be == nil {
		t.Fatalf("expected positive match for /api/v1/orders/1234")
	}
	if be.Service != "orders-svc" || be.Port != 8080 {
		t.Fatalf("expected orders-svc:8080, got %s:%d", be.Service, be.Port)
	}

	// Match negative path
	_, _, negativeMatched := mgr.MatchRoute(gwKey, "/static/images/logo.png", nil)
	if negativeMatched {
		t.Fatalf("expected non-match for /static/images/logo.png")
	}
}

func TestReconcileGatewayWithNilHostnameListener(t *testing.T) {
	log := zap.NewNop()
	scheme := runtime.NewScheme()
	_ = gwv1.Install(scheme)

	gwObj := &gwv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "nil-host-gw",
		},
		Spec: gwv1.GatewaySpec{
			GatewayClassName: "strait-class",
			Listeners: []gwv1.Listener{
				{
					Name:     "http",
					Port:     80,
					Protocol: gwv1.HTTPProtocolType,
					Hostname: nil, // Must not panic
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(gwObj).Build()
	mgr := NewManager(client, log)
	ctx := context.Background()

	gwKey := types.NamespacedName{Namespace: "default", Name: "nil-host-gw"}
	if err := mgr.ReconcileGateway(ctx, gwKey); err != nil {
		t.Fatalf("reconcile gateway with nil hostname failed: %v", err)
	}

	all := mgr.GetAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 gateway, got %d", len(all))
	}
	if len(all[0].Listeners) != 1 || all[0].Listeners[0].Hostname != "" {
		t.Fatalf("expected listener hostname to be empty string, got %q", all[0].Listeners[0].Hostname)
	}
}
