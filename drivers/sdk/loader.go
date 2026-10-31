package sdk

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ProfileRegistry manages the collection of known device profiles.
type ProfileRegistry struct {
	mu       sync.RWMutex
	profiles map[string]*DeviceProfile
}

// NewProfileRegistry initializes an empty profile registry.
func NewProfileRegistry() *ProfileRegistry {
	return &ProfileRegistry{
		profiles: make(map[string]*DeviceProfile),
	}
}

var DefaultRegistry = NewProfileRegistry()

// Register adds or replaces a profile in the registry.
func (r *ProfileRegistry) Register(p *DeviceProfile) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profiles[p.Name] = p
}

// Get looks up a profile by name.
func (r *ProfileRegistry) Get(name string) (*DeviceProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.profiles[name]
	return p, ok
}

// FindForSysObjectID attempts to match a vendor profile by SNMP sysObjectID.
func (r *ProfileRegistry) FindForSysObjectID(oid string) (*DeviceProfile, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.profiles {
		if p.SysObjectIDPrefix != "" && strings.HasPrefix(oid, p.SysObjectIDPrefix) {
			return p, true
		}
	}
	return nil, false
}

// List returns a slice of all registered profiles.
func (r *ProfileRegistry) List() []*DeviceProfile {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*DeviceProfile, 0, len(r.profiles))
	for _, p := range r.profiles {
		list = append(list, p)
	}
	return list
}

// ParseProfile deserializes a YAML profile payload.
func ParseProfile(data []byte) (*DeviceProfile, error) {
	var p DeviceProfile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing device profile: %w", err)
	}
	if p.Name == "" {
		return nil, fmt.Errorf("device profile missing required 'name' field")
	}
	return &p, nil
}

// LoadFile reads and registers a profile from a YAML file.
func (r *ProfileRegistry) LoadFile(path string) (*DeviceProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := ParseProfile(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	r.Register(p)
	return p, nil
}

// LoadFS reads all .yaml and .yml profiles from an fs.FS filesystem.
func (r *ProfileRegistry) LoadFS(sys fs.FS, root string) error {
	return fs.WalkDir(sys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}
		data, err := fs.ReadFile(sys, path)
		if err != nil {
			return err
		}
		p, err := ParseProfile(data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		r.Register(p)
		return nil
	})
}
