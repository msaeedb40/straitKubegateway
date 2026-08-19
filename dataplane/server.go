// Package dataplane provides the CNI daemon socket server that bridges
// the short-lived CNI plugin binary to the long-running straitd dataplane manager.
package dataplane

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

const (
	// DefaultSocketPath is the default UNIX socket path for CNI-to-daemon IPC.
	DefaultSocketPath = "/var/run/strait/cni.sock"
)

// CNIRequest is the wire format sent by the CNI plugin over the UNIX socket.
type CNIRequest struct {
	Command     string            `json:"command"` // "ADD" or "DEL"
	ContainerID string            `json:"container_id"`
	NetnsPath   string            `json:"netns_path"`
	IfName      string            `json:"if_name"`
	Namespace   string            `json:"namespace"`
	PodName     string            `json:"pod_name"`
	SegmentID   uint32            `json:"segment_id"`
	Labels      map[string]string `json:"labels"`
	MTU         int               `json:"mtu"`
	PodCIDR     string            `json:"pod_cidr"`
}

// CNIResponse is the wire format returned to the CNI plugin.
type CNIResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	IP        string `json:"ip,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	PrefixLen int    `json:"prefix_len,omitempty"`
	HostVeth  string `json:"host_veth,omitempty"`
	IfIndex   int    `json:"if_index,omitempty"`
}

// Server listens on a UNIX domain socket and dispatches CNI requests
// to the persistent dataplane Manager.
type Server struct {
	manager    *Manager
	socketPath string
	listener   net.Listener
	logger     *logging.Logger
}

// NewServer creates a new CNI daemon socket server backed by the given Manager.
func NewServer(mgr *Manager, socketPath string) *Server {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	return &Server{
		manager:    mgr,
		socketPath: socketPath,
		logger:     logging.DefaultLogger(),
	}
}

// Start begins listening on the UNIX socket and handling connections.
// It removes any stale socket file before binding.
func (s *Server) Start() error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(s.socketPath), 0o755); err != nil {
		return fmt.Errorf("create socket directory: %w", err)
	}

	// Remove stale socket from previous run
	if err := os.RemoveAll(s.socketPath); err != nil {
		return fmt.Errorf("remove stale socket: %w", err)
	}

	ln, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", s.socketPath, err)
	}
	s.listener = ln

	// Allow non-root CNI plugin processes to connect
	if err := os.Chmod(s.socketPath, 0o666); err != nil {
		s.logger.Error("failed to chmod socket", &types.ErrorInfo{
			Code: "SOCK_CHMOD_ERR", Message: err.Error(),
		}, &types.Metadata{Component: "cni-server"})
	}

	s.logger.Info(fmt.Sprintf("CNI daemon socket listening on %s", s.socketPath),
		&types.Metadata{Component: "cni-server"})

	go s.acceptLoop()
	return nil
}

// Stop closes the listener and removes the socket file.
func (s *Server) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	_ = os.RemoveAll(s.socketPath)
}

// acceptLoop accepts incoming connections and handles each in a goroutine.
func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			// Listener was closed (shutdown)
			return
		}
		go s.handleConnection(conn)
	}
}

// handleConnection reads a single CNIRequest, dispatches to the Manager,
// and writes back a CNIResponse.
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	var req CNIRequest
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&req); err != nil {
		s.writeError(conn, fmt.Sprintf("decode request: %v", err))
		return
	}

	meta := &types.Metadata{
		Component: "cni-server",
		Namespace: req.Namespace,
	}

	switch req.Command {
	case "ADD":
		s.handleAdd(conn, &req, meta)
	case "DEL":
		s.handleDel(conn, &req, meta)
	default:
		s.writeError(conn, fmt.Sprintf("unknown command: %s", req.Command))
	}
}

// handleAdd processes a CNI ADD request through the persistent dataplane Manager.
func (s *Server) handleAdd(conn net.Conn, req *CNIRequest, meta *types.Metadata) {
	result, err := s.manager.AddPodNetwork(PodNetworkRequest{
		ContainerID: req.ContainerID,
		NetnsPath:   req.NetnsPath,
		IfName:      req.IfName,
		Namespace:   req.Namespace,
		PodName:     req.PodName,
		SegmentID:   types.SegmentID(req.SegmentID),
		Labels:      req.Labels,
		MTU:         req.MTU,
	})
	if err != nil {
		s.logger.Error(fmt.Sprintf("CNI ADD failed for %s/%s: %v", req.Namespace, req.PodName, err),
			&types.ErrorInfo{Code: "CNI_ADD_ERR", Message: err.Error()}, meta)
		s.writeError(conn, err.Error())
		return
	}

	resp := CNIResponse{
		Success:   true,
		IP:        result.IP.String(),
		Gateway:   result.Gateway.String(),
		PrefixLen: result.PrefixLen,
		HostVeth:  result.HostVeth,
		IfIndex:   result.IfIndex,
	}

	s.logger.Info(fmt.Sprintf("CNI ADD success: %s/%s -> %s", req.Namespace, req.PodName, result.IP), meta)
	s.writeResponse(conn, &resp)
}

// handleDel processes a CNI DEL request through the persistent dataplane Manager.
func (s *Server) handleDel(conn net.Conn, req *CNIRequest, meta *types.Metadata) {
	err := s.manager.DelPodNetwork(req.ContainerID)
	if err != nil {
		s.logger.Error(fmt.Sprintf("CNI DEL failed for container %s: %v", req.ContainerID, err),
			&types.ErrorInfo{Code: "CNI_DEL_ERR", Message: err.Error()}, meta)
		s.writeError(conn, err.Error())
		return
	}

	s.logger.Info(fmt.Sprintf("CNI DEL success: container %s", req.ContainerID), meta)
	s.writeResponse(conn, &CNIResponse{Success: true})
}

func (s *Server) writeResponse(conn net.Conn, resp *CNIResponse) {
	_ = json.NewEncoder(conn).Encode(resp)
}

func (s *Server) writeError(conn net.Conn, msg string) {
	_ = json.NewEncoder(conn).Encode(&CNIResponse{
		Success: false,
		Error:   msg,
	})
}
