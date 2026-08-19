package gateway

import (
	"fmt"
	"sync"
)

// Manager coordinates all active Gateway API instances and routes.
type Manager struct {
	mu       sync.RWMutex
	gateways map[string]*GatewayInstance // "namespace/name" -> GatewayInstance
	router   *Router
}

// NewManager creates a new Gateway manager.
func NewManager() *Manager {
	return &Manager{
		gateways: make(map[string]*GatewayInstance),
		router:   NewRouter(nil),
	}
}

// UpsertGateway adds or updates a Gateway instance.
func (m *Manager) UpsertGateway(gw GatewayInstance) *GatewayInstance {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := gw.Namespace + "/" + gw.Name
	if gw.Routes == nil {
		gw.Routes = make(map[string]*Route)
	}

	m.gateways[key] = &gw
	m.rebuildRouterLocked()
	return &gw
}

// DeleteGateway removes a Gateway instance.
func (m *Manager) DeleteGateway(ns, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := ns + "/" + name
	delete(m.gateways, key)
	m.rebuildRouterLocked()
}

// UpsertRoute binds a Route to a Gateway instance.
func (m *Manager) UpsertRoute(gwNs, gwName string, route Route) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gwKey := gwNs + "/" + gwName
	gw, exists := m.gateways[gwKey]
	if !exists {
		return fmt.Errorf("gateway %s/%s not found", gwNs, gwName)
	}

	routeKey := route.Namespace + "/" + route.Name
	gw.Routes[routeKey] = &route

	m.rebuildRouterLocked()
	return nil
}

// DeleteRoute unbinds a Route from a Gateway instance.
func (m *Manager) DeleteRoute(gwNs, gwName, routeNs, routeName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	gwKey := gwNs + "/" + gwName
	if gw, exists := m.gateways[gwKey]; exists {
		routeKey := routeNs + "/" + routeName
		delete(gw.Routes, routeKey)
		m.rebuildRouterLocked()
	}
}

// Router returns the active compiled Router.
func (m *Manager) Router() *Router {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.router
}

func (m *Manager) rebuildRouterLocked() {
	var allRoutes []*Route
	for _, gw := range m.gateways {
		for _, r := range gw.Routes {
			allRoutes = append(allRoutes, r)
		}
	}
	m.router = NewRouter(allRoutes)
}
