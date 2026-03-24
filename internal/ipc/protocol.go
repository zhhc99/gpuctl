package ipc

import "github.com/zhhc99/gpuctl/internal/gpu"

type FieldResult struct {
	Field   string `json:"field"`
	Skipped bool   `json:"skipped,omitempty"`
	Err     string `json:"error,omitempty"`
}

type DeviceResult struct {
	Index  int           `json:"index"`
	Name   string        `json:"name"`
	Fields []FieldResult `json:"fields"`
}

type LoadResponse struct {
	Devices []DeviceResult `json:"devices"`
	Err     string         `json:"error,omitempty"`
}

type ListRequest struct {
	All     bool  `json:"all,omitempty"`
	Indices []int `json:"indices,omitempty"`
}

type ListResponse struct {
	Devices []gpu.Snapshot `json:"devices"`
	Err     string         `json:"error,omitempty"`
}

// TuneGetRequest selects devices; all parameters are always returned.
type TuneGetRequest struct {
	All     bool  `json:"all,omitempty"`
	Indices []int `json:"indices,omitempty"`
}

type SpecRow struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Unit    string `json:"unit"`
	Current string `json:"current"`
	Default string `json:"default"`
	Min     string `json:"min"`
	Max     string `json:"max"`
}

type DeviceSpec struct {
	Index int       `json:"index"`
	Name  string    `json:"name"`
	Rows  []SpecRow `json:"rows"`
}

type TuneGetResponse struct {
	Devices []DeviceSpec `json:"devices"`
	Err     string       `json:"error,omitempty"`
}

type TuneSetRequest struct {
	All     bool           `json:"all,omitempty"`
	Indices []int          `json:"indices,omitempty"`
	Updates map[string]int `json:"updates"`
}

type TuneSetResponse struct {
	Devices []DeviceResult `json:"devices"`
	Err     string         `json:"error,omitempty"`
}

type TuneResetRequest struct {
	All     bool     `json:"all,omitempty"`
	Indices []int    `json:"indices,omitempty"`
	Keys    []string `json:"keys,omitempty"`
}

type TuneResetResponse struct {
	Devices []DeviceResult `json:"devices"`
	Err     string         `json:"error,omitempty"`
}

type VersionResponse struct {
	BackendName    string `json:"backend_name"`
	BackendVersion string `json:"backend_version"`
	DriverVersion  string `json:"driver_version"`
	Err            string `json:"error,omitempty"`
}
