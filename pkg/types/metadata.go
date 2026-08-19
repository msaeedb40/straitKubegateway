package types

import "time"

// Metadata is the common observability metadata envelope carried by every
// significant operation, event, log entry, metric, and trace span in
// straitKubegateway. It enables rich correlation across subsystems.
type Metadata struct {
	// Timestamp is the UTC time of the event or operation.
	Timestamp time.Time `json:"timestamp"`

	// Service identification
	ServiceName    string `json:"service,omitempty"`
	ServiceVersion string `json:"version,omitempty"`
	Component      string `json:"component,omitempty"`

	// Environment context
	Environment string `json:"environment,omitempty"`
	Region      string `json:"region,omitempty"`
	Zone        string `json:"zone,omitempty"`

	// Infrastructure context
	ClusterID      string `json:"cluster_id,omitempty"`
	ClusterName    string `json:"cluster_name,omitempty"`
	ClusterVersion string `json:"cluster_version,omitempty"`
	NodeID         string `json:"node_id,omitempty"`
	NodeName       string `json:"node_name,omitempty"`
	NodeIP         string `json:"node_ip,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	PodName        string `json:"pod,omitempty"`
	PodID          string `json:"pod_id,omitempty"`
	ContainerName  string `json:"container,omitempty"`
	ContainerID    string `json:"container_id,omitempty"`

	// Gateway/network context
	GatewayID  string `json:"gateway_id,omitempty"`
	TunnelID   string `json:"tunnel_id,omitempty"`
	FlowID     string `json:"flow_id,omitempty"`
	RouteID    string `json:"route_id,omitempty"`
	PolicyID   string `json:"policy_id,omitempty"`
	SegmentID  string `json:"segment_id,omitempty"`
	ServiceID  string `json:"service_id,omitempty"`
	EndpointID string `json:"endpoint_id,omitempty"`

	// Distributed tracing
	TraceID   string `json:"trace_id,omitempty"`
	SpanID    string `json:"span_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	CommandID string `json:"command_id,omitempty"`

	// Operation context
	Operation string `json:"operation,omitempty"`
	Action    string `json:"action,omitempty"`
	Status    string `json:"status,omitempty"`

	// Duration
	DurationMs float64 `json:"duration_ms,omitempty"`
}

// LatencyStats holds percentile latency measurements.
type LatencyStats struct {
	Unit  string  `json:"unit"`
	Count int64   `json:"count"`
	Min   float64 `json:"min"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
	Max   float64 `json:"max"`
}

// LatencyBreakdown provides a request latency decomposition.
type LatencyBreakdown struct {
	TotalMs         float64 `json:"total_ms"`
	AuthMs          float64 `json:"auth_ms,omitempty"`
	ValidationMs    float64 `json:"validation_ms,omitempty"`
	QueueMs         float64 `json:"queue_ms,omitempty"`
	ControllerMs    float64 `json:"controller_ms,omitempty"`
	KubernetesMs    float64 `json:"kubernetes_ms,omitempty"`
	SerializationMs float64 `json:"serialization_ms,omitempty"`
	NetworkMs       float64 `json:"network_ms,omitempty"`
}

// TrafficStats holds throughput and packet counters.
type TrafficStats struct {
	RxBytesPerSecond   uint64 `json:"rx_bytes_per_second"`
	TxBytesPerSecond   uint64 `json:"tx_bytes_per_second"`
	RxPacketsPerSecond uint64 `json:"rx_per_second"`
	TxPacketsPerSecond uint64 `json:"tx_per_second"`
	DropsPerSecond     uint64 `json:"drops_per_second"`
}

// HealthStatus represents the health of a component.
type HealthStatus struct {
	Availability float64 `json:"availability"`
	ErrorRate    float64 `json:"error_rate"`
}

// ErrorInfo provides machine-readable error context.
type ErrorInfo struct {
	Code      string `json:"code"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
