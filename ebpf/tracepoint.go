package ebpf

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

// TracepointManager manages kernel tracepoints and perf/ringbuf event streams.
type TracepointManager struct {
	mu     sync.RWMutex
	links  []link.Link
	reader *ringbuf.Reader
}

// NewTracepointManager creates a new tracepoint manager.
func NewTracepointManager() *TracepointManager {
	return &TracepointManager{
		links: make([]link.Link, 0),
	}
}

// AttachTracepoint attaches a BPF program to a kernel tracepoint.
func (m *TracepointManager) AttachTracepoint(group, name string, prog *ebpf.Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := link.Tracepoint(group, name, prog, nil)
	if err != nil {
		return fmt.Errorf("attach tracepoint %s/%s: %w", group, name, err)
	}

	m.links = append(m.links, l)
	return nil
}

// StartEventReader starts reading flow events from a ring buffer map.
func (m *TracepointManager) StartEventReader(ctx context.Context, eventsMap *ebpf.Map, handler func([]byte)) error {
	rd, err := ringbuf.NewReader(eventsMap)
	if err != nil {
		return fmt.Errorf("open ringbuf reader: %w", err)
	}

	m.mu.Lock()
	m.reader = rd
	m.mu.Unlock()

	go func() {
		defer rd.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				record, err := rd.Read()
				if err != nil {
					if errors.Is(err, ringbuf.ErrClosed) {
						return
					}
					continue
				}
				if handler != nil {
					handler(record.RawSample)
				}
			}
		}
	}()

	return nil
}

// Close closes all attached tracepoints and ring buffer readers.
func (m *TracepointManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.reader != nil {
		_ = m.reader.Close()
		m.reader = nil
	}

	for _, l := range m.links {
		_ = l.Close()
	}
	m.links = nil
	return nil
}
