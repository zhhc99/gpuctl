package gpu

type Backend interface {
	Name() string    // "NVML"/"OneAPI"/...
	Version() string // version of NVML/OneAPI/...
	DriverVersion() string

	Init() error
	Shutdown() error
	GPUs() ([]Device, error)
}

type BackendInfo struct {
	ManagerName    string
	ManagerVersion string
	DriverVersion  string
}

func (m *BackendInfo) Capture(mgr Backend) {
	m.ManagerName = mgr.Name()
	m.ManagerVersion = mgr.Version()
	m.DriverVersion = mgr.DriverVersion()
}
