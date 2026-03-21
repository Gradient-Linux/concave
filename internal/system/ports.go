package system

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Gradient-Linux/concave/internal/config"
	"github.com/Gradient-Linux/concave/internal/suite"
)

// PortConflict describes a detected suite port collision.
type PortConflict struct {
	Port            int
	ExistingSuite   string
	ExistingService string
	NewSuite        string
	NewService      string
}

type portRegistration struct {
	Suite   string
	Service string
}

var (
	portRegistryMu sync.Mutex
	portRegistry   = map[int]portRegistration{}
)

// CheckConflicts returns any port conflicts for the provided suite.
func CheckConflicts(s suite.Suite) []PortConflict {
	portRegistryMu.Lock()
	defer portRegistryMu.Unlock()

	registry := loadPersistedRegistry()
	for port, registration := range portRegistry {
		registry[port] = registration
	}
	return checkConflicts(s, registry)
}

// Register adds suite ports to the runtime registry.
func Register(s suite.Suite) error {
	portRegistryMu.Lock()
	defer portRegistryMu.Unlock()

	registry := loadPersistedRegistry()
	for port, registration := range portRegistry {
		registry[port] = registration
	}
	for _, conflict := range checkConflicts(s, registry) {
		return fmt.Errorf("port %d already used by %s (%s)", conflict.Port, conflict.ExistingSuite, conflict.ExistingService)
	}

	for _, mapping := range s.Ports {
		if existing, ok := portRegistry[mapping.Port]; ok && sharedMLflow(existing.Suite, s.Name, mapping.Port) {
			continue
		}
		portRegistry[mapping.Port] = portRegistration{
			Suite:   s.Name,
			Service: mapping.Service,
		}
	}
	return nil
}

// Deregister removes a suite from the runtime registry.
func Deregister(s suite.Suite) error {
	portRegistryMu.Lock()
	defer portRegistryMu.Unlock()

	for _, mapping := range s.Ports {
		if existing, ok := portRegistry[mapping.Port]; ok && existing.Suite == s.Name {
			delete(portRegistry, mapping.Port)
		}
	}
	return nil
}

func loadPersistedRegistry() map[int]portRegistration {
	registry := map[int]portRegistration{}
	state, err := config.LoadState()
	if err != nil {
		return registry
	}
	for _, name := range state.Installed {
		s, err := suite.Get(name)
		if err != nil {
			continue
		}
		for _, mapping := range s.Ports {
			if existing, ok := registry[mapping.Port]; ok && sharedMLflow(existing.Suite, s.Name, mapping.Port) {
				continue
			}
			registry[mapping.Port] = portRegistration{Suite: s.Name, Service: mapping.Service}
		}
	}
	return registry
}

func checkConflicts(s suite.Suite, registry map[int]portRegistration) []PortConflict {
	var conflicts []PortConflict
	for _, mapping := range s.Ports {
		if existing, ok := registry[mapping.Port]; ok {
			if sharedMLflow(existing.Suite, s.Name, mapping.Port) {
				continue
			}
			conflicts = append(conflicts, PortConflict{
				Port:            mapping.Port,
				ExistingSuite:   existing.Suite,
				ExistingService: existing.Service,
				NewSuite:        s.Name,
				NewService:      mapping.Service,
			})
		}
	}
	return conflicts
}

func sharedMLflow(a, b string, port int) bool {
	if port != 5000 {
		return false
	}
	key := strings.Join([]string{a, b}, ":")
	return key == "boosting:flow" || key == "flow:boosting"
}
