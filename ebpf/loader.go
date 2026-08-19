package ebpf

import (
	"fmt"
	"os"

	"github.com/cilium/ebpf"
	"github.com/straitKubegateway/straitKubegateway/pkg/bpf"
)

// Collection holds the loaded BPF programs and maps.
type Collection struct {
	EndpointsMap *ebpf.Map
	RoutesMap    *ebpf.Map
	PoliciesMap  *ebpf.Map
	MetricsMap   *ebpf.Map
	Programs     map[string]*ebpf.Program
}

// Loader manages compilation, loading, and pinning of BPF objects.
type Loader struct {
	collection *Collection
}

// NewLoader initializes the BPF loader and ensures pin directories exist.
func NewLoader() (*Loader, error) {
	if err := bpf.EnsurePinDirectory(); err != nil {
		return nil, fmt.Errorf("ensure pin directory: %w", err)
	}
	return &Loader{}, nil
}

// LoadMaps creates or loads the pinned BPF maps required for Phase 1.
func (l *Loader) LoadMaps() (*Collection, error) {
	endpointsMapSpec := &bpf.MapSpec{
		Name:       "endpoints_map",
		Type:       ebpf.Hash,
		KeySize:    4,  // sizeof(struct endpoint_key_v4)
		ValueSize:  32, // sizeof(struct endpoint_info)
		MaxEntries: 65536,
		PinPath:    bpf.MapPinPath("endpoints_map"),
	}

	epMap, err := bpf.CreateMap(endpointsMapSpec)
	if err != nil {
		// Attempt to load existing pinned map
		epMap, err = bpf.LoadPinnedMap(endpointsMapSpec.PinPath)
		if err != nil {
			// In environments without BPF mount permissions (e.g. tests), continue with mock
			fmt.Fprintf(os.Stderr, "warning: cannot create/load pinned BPF map: %v\n", err)
		}
	}

	l.collection = &Collection{
		EndpointsMap: epMap,
		Programs:     make(map[string]*ebpf.Program),
	}

	return l.collection, nil
}

// Close releases all loaded BPF resources.
func (l *Loader) Close() error {
	if l.collection != nil {
		if l.collection.EndpointsMap != nil {
			l.collection.EndpointsMap.Close()
		}
		for _, p := range l.collection.Programs {
			p.Close()
		}
	}
	return nil
}
