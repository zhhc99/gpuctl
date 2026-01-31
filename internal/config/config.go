package config

import (
	"encoding/json"
	"os"
	"sync"
)

const DefaultFileName = "config.json"

type GpuSettings struct {
	PowerLimit     int `json:"power_limit"`      // W
	ClockOffsetGPU int `json:"clock_offset_gpu"` // MHz
	ClockOffsetMem int `json:"clock_offset_mem"` // MHz
	ClockLimitGPU  int `json:"clock_limit_gpu"`  // MHz
}

type Manager struct {
	mu       sync.Mutex
	FilePath string
	Settings map[string]GpuSettings // Key: GPU UUID
}

func New(path string) *Manager {
	if path == "" {
		path = DefaultFileName
	}
	return &Manager{
		FilePath: path,
		Settings: make(map[string]GpuSettings),
	}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.FilePath)
	if os.IsNotExist(err) {
		return nil // No config yet, that's fine
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.Settings)
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(m.Settings, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.FilePath, data, 0644)
}

func (m *Manager) Get(uuid string) (GpuSettings, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.Settings[uuid]
	return s, ok
}

func (m *Manager) Set(uuid string, s GpuSettings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Settings[uuid] = s
}
