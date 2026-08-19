package policy

import (
	"sync"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
)

// Manager manages active StraitNetworkPolicy objects and provides fast evaluation.
type Manager struct {
	mu        sync.RWMutex
	compiler  *Compiler
	policies  map[string]*v1alpha1.StraitNetworkPolicy
	evaluator *Evaluator
}

// NewManager creates a new policy manager.
func NewManager() *Manager {
	return &Manager{
		compiler:  NewCompiler(),
		policies:  make(map[string]*v1alpha1.StraitNetworkPolicy),
		evaluator: NewEvaluator(nil),
	}
}

// UpsertPolicy adds or updates a policy and recomputes the compiled rule set.
func (m *Manager) UpsertPolicy(p *v1alpha1.StraitNetworkPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := p.Namespace + "/" + p.Name
	if p.Namespace == "" {
		key = p.Name
	}
	m.policies[key] = p
	m.recompileLocked()
}

// DeletePolicy removes a policy and recomputes the compiled rule set.
func (m *Manager) DeletePolicy(ns, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := ns + "/" + name
	if ns == "" {
		key = name
	}
	delete(m.policies, key)
	m.recompileLocked()
}

// Evaluator returns the current active policy evaluator.
func (m *Manager) Evaluator() *Evaluator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.evaluator
}

func (m *Manager) recompileLocked() {
	var allRules []CompiledRule
	for _, p := range m.policies {
		rules := m.compiler.Compile(p, nil, nil)
		allRules = append(allRules, rules...)
	}
	m.evaluator = NewEvaluator(allRules)
}
