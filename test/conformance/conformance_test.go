// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package conformance

import (
	"testing"

	"github.com/straitkubegateway/straitkubegateway/pkg/types"
)

func TestReadinessConditionInvariants(t *testing.T) {
	// Invariant: CNIReady, ServiceReady, PolicyReady, GatewayReady, TransitReady are independent
	conditions := []string{
		types.ConditionCNIReady,
		types.ConditionServiceReady,
		types.ConditionPolicyReady,
		types.ConditionGatewayReady,
		types.ConditionTransitReady,
		types.ConditionBGPReady,
	}

	seen := make(map[string]bool)
	for _, c := range conditions {
		if seen[c] {
			t.Errorf("duplicate condition key %s", c)
		}
		seen[c] = true
	}

	if len(seen) != 6 {
		t.Errorf("expected 6 distinct readiness conditions, got %d", len(seen))
	}
}
