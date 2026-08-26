// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package observability implements Prometheus metrics, OpenTelemetry tracing,
// structured logging, eBPF flow visibility, and health monitoring.
//
// Common metadata model: every major object carries these fields:
// cluster-id, node-id, namespace, pod, service, endpoint, flow-id,
// trace-id, policy-id, segment-id, gateway-id.
//
// Used across: logs, metrics, traces, flow events, policy decisions,
// BGP events, transit events, CNI events.
package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Common metadata model
// ============================================================================

// Metadata is the canonical observability metadata attached to every
// major event, log entry, metric label set, trace span, and flow event.
// Invariant: every major object carries ALL of these fields.
type Metadata struct {
	ClusterID  sgtypes.ClusterID `json:"clusterID"`
	NodeID     sgtypes.NodeID    `json:"nodeID"`
	Namespace  string            `json:"namespace,omitempty"`
	Pod        string            `json:"pod,omitempty"`
	Service    string            `json:"service,omitempty"`
	Endpoint   string            `json:"endpoint,omitempty"`
	FlowID     string            `json:"flowID,omitempty"`
	TraceID    string            `json:"traceID,omitempty"`
	PolicyID   string            `json:"policyID,omitempty"`
	SegmentID  sgtypes.SegmentID `json:"segmentID"`
	GatewayID  string            `json:"gatewayID,omitempty"`
}

// ZapFields converts the metadata to zap log fields.
func (m Metadata) ZapFields() []zap.Field {
	return []zap.Field{
		zap.String("clusterID", string(m.ClusterID)),
		zap.String("nodeID", string(m.NodeID)),
		zap.String("namespace", m.Namespace),
		zap.String("pod", m.Pod),
		zap.String("service", m.Service),
		zap.String("endpoint", m.Endpoint),
		zap.String("flowID", m.FlowID),
		zap.String("traceID", m.TraceID),
		zap.String("policyID", m.PolicyID),
		zap.Uint32("segmentID", uint32(m.SegmentID)),
		zap.String("gatewayID", m.GatewayID),
	}
}

// PrometheusLabels returns the metadata as Prometheus label values.
func (m Metadata) PrometheusLabels() prometheus.Labels {
	return prometheus.Labels{
		"cluster_id":  string(m.ClusterID),
		"node_id":     string(m.NodeID),
		"namespace":   m.Namespace,
		"pod":         m.Pod,
		"service":     m.Service,
		"segment_id":  m.SegmentID.String(),
		"gateway_id":  m.GatewayID,
	}
}

// ============================================================================
// Prometheus Metrics
// ============================================================================

// Metrics holds all straitKubegateway Prometheus metrics.
type Metrics struct {
	// CNI metrics
	CNIAddTotal    prometheus.Counter
	CNIDelTotal    prometheus.Counter
	CNIAddDuration prometheus.Histogram
	CNIErrors      prometheus.Counter

	// Service LB metrics
	ServiceTotal    prometheus.Gauge
	BackendTotal    prometheus.Gauge
	LBPacketsTotal  prometheus.Counter

	// Policy metrics
	PolicyTotal      prometheus.Gauge
	PolicyDropsTotal prometheus.Counter

	// NAT metrics
	NATRulesTotal     prometheus.Gauge
	ConntrackTotal    prometheus.Gauge

	// Transit metrics
	TransitSegmentsTotal prometheus.Gauge
	TransitPeersTotal    prometheus.Gauge

	// BGP metrics
	BGPPeersTotal   prometheus.Gauge
	BGPRoutesTotal  prometheus.Gauge

	// Node metrics
	IdentitiesTotal prometheus.Gauge
}

const namespace = "straitkubegateway"

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		CNIAddTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Name: "cni_add_total",
			Help: "Total number of CNI ADD operations.",
		}),
		CNIDelTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Name: "cni_del_total",
			Help: "Total number of CNI DEL operations.",
		}),
		CNIAddDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace, Name: "cni_add_duration_seconds",
			Help:    "Duration of CNI ADD operations in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
		}),
		CNIErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Name: "cni_errors_total",
			Help: "Total number of CNI errors.",
		}),
		ServiceTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "services_total",
			Help: "Number of programmed Kubernetes services.",
		}),
		BackendTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "service_backends_total",
			Help: "Number of programmed service backends.",
		}),
		LBPacketsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Name: "lb_packets_total",
			Help: "Total packets load-balanced by the eBPF service LB.",
		}),
		PolicyTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "policies_total",
			Help: "Number of compiled network policies.",
		}),
		PolicyDropsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace, Name: "policy_drops_total",
			Help: "Total packets dropped by the policy engine.",
		}),
		NATRulesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "nat_rules_total",
			Help: "Number of programmed NAT rules.",
		}),
		ConntrackTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "conntrack_entries_total",
			Help: "Number of active conntrack entries.",
		}),
		TransitSegmentsTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "transit_segments_total",
			Help: "Number of active transit segments.",
		}),
		TransitPeersTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "transit_peers_total",
			Help: "Number of connected transit peers.",
		}),
		BGPPeersTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "bgp_peers_total",
			Help: "Number of established BGP peers.",
		}),
		BGPRoutesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "bgp_routes_total",
			Help: "Number of BGP routes in the RIB.",
		}),
		IdentitiesTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace, Name: "identities_total",
			Help: "Number of allocated BPF identities.",
		}),
	}

	// Register all metrics
	reg.MustRegister(
		m.CNIAddTotal, m.CNIDelTotal, m.CNIAddDuration, m.CNIErrors,
		m.ServiceTotal, m.BackendTotal, m.LBPacketsTotal,
		m.PolicyTotal, m.PolicyDropsTotal,
		m.NATRulesTotal, m.ConntrackTotal,
		m.TransitSegmentsTotal, m.TransitPeersTotal,
		m.BGPPeersTotal, m.BGPRoutesTotal,
		m.IdentitiesTotal,
	)
	return m
}

// ============================================================================
// Metrics server
// ============================================================================

// Server exposes Prometheus metrics over HTTP.
type Server struct {
	addr    string
	handler http.Handler
	log     *zap.Logger
}

// NewServer creates a metrics HTTP server on the given address.
func NewServer(addr string, reg *prometheus.Registry, log *zap.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})
	return &Server{addr: addr, handler: mux, log: log}
}

// Serve starts the metrics server. Blocks until ctx is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()
	s.log.Info("metrics server starting", zap.String("addr", s.addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("metrics server: %w", err)
	}
	return nil
}

// ============================================================================
// Structured logger
// ============================================================================

// NewLogger creates a production zap logger.
func NewLogger(debug bool) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.Development = true
	}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build()
}

// ============================================================================
// Flow event
// ============================================================================

// FlowDirection is the direction of a network flow.
type FlowDirection string

const (
	FlowDirectionIngress FlowDirection = "ingress"
	FlowDirectionEgress  FlowDirection = "egress"
)

// FlowEvent represents a network flow visibility event from eBPF.
// Used only by observability hooks (tracepoints/kprobes/perf/ringbuf).
// NOT used for packet forwarding.
type FlowEvent struct {
	Metadata  Metadata
	Direction FlowDirection
	SrcIP     string
	DstIP     string
	SrcPort   uint16
	DstPort   uint16
	Proto     uint8
	Bytes     uint64
	Packets   uint64
	Action    string // allowed, denied, rejected
	Timestamp time.Time
}
