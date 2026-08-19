// Package flow provides the eBPF network flow observation and aggregation pipeline.
package flow

import (
	"sync"
	"time"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Record represents a single network flow record captured by the eBPF dataplane.
type Record struct {
	ID             string          `json:"id"`
	Timestamp      time.Time       `json:"timestamp"`
	SourceIP       string          `json:"source_ip"`
	DestIP         string          `json:"dest_ip"`
	SourcePort     uint16          `json:"source_port"`
	DestPort       uint16          `json:"dest_port"`
	Protocol       types.Protocol  `json:"protocol"`
	Direction      types.Direction `json:"direction"`
	Namespace      string          `json:"namespace"`
	PodName        string          `json:"pod_name"`
	GatewayID      string          `json:"gateway_id,omitempty"`
	TunnelID       string          `json:"tunnel_id,omitempty"`
	SegmentID      types.SegmentID `json:"segment_id"`
	PolicyDecision string          `json:"policy_decision"` // ALLOW, DENY, REJECT
	Bytes          uint64          `json:"bytes"`
	Packets        uint64          `json:"packets"`
	DurationMs     float64         `json:"duration_ms"`
	TraceID        string          `json:"trace_id,omitempty"`
}

// Pipeline aggregates and buffers flow records before streaming to observers.
type Pipeline struct {
	mu      sync.RWMutex
	records []Record
	maxSize int
}

// NewPipeline creates a new flow pipeline with the specified ring buffer size.
func NewPipeline(maxSize int) *Pipeline {
	if maxSize <= 0 {
		maxSize = 10000
	}
	return &Pipeline{
		records: make([]Record, 0, maxSize),
		maxSize: maxSize,
	}
}

// Ingest adds a new flow record to the buffer.
func (p *Pipeline) Ingest(r Record) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.records) >= p.maxSize {
		// Drop oldest 10%
		drop := p.maxSize / 10
		if drop < 1 {
			drop = 1
		}
		p.records = p.records[drop:]
	}
	p.records = append(p.records, r)
}

// Query returns records matching the given criteria.
func (p *Pipeline) Query(namespace string, segmentID types.SegmentID, limit int) []Record {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if limit <= 0 || limit > len(p.records) {
		limit = len(p.records)
	}

	result := make([]Record, 0, limit)
	for i := len(p.records) - 1; i >= 0 && len(result) < limit; i-- {
		r := p.records[i]
		if namespace != "" && r.Namespace != namespace {
			continue
		}
		if segmentID != types.SegmentBackbone && r.SegmentID != segmentID {
			continue
		}
		result = append(result, r)
	}
	return result
}
