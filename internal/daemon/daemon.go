package daemon

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zhhc99/gpuctl/internal/config"
	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
	"github.com/zhhc99/gpuctl/internal/nvml"
)

var (
	backend gpu.Backend
	devices []gpu.Device
)

func Run() error {
	return runDaemon()
}

func initBackend() error {
	var err error
	backend, err = nvml.NewBackend()
	if err != nil {
		return fmt.Errorf("nvml unavailable: %w", err)
	}
	if err := backend.Init(); err != nil {
		return fmt.Errorf("nvml init failed: %w", err)
	}
	devs, err := backend.GPUs()
	if err != nil {
		return fmt.Errorf("failed to list GPUs: %w", err)
	}
	sort.Slice(devs, func(i, j int) bool {
		return devs[i].PCIBusID() < devs[j].PCIBusID()
	})
	devices = devs
	return nil
}

func printDeviceResults(results []ipc.DeviceResult) {
	for _, dr := range results {
		fmt.Printf(locale.T("msg.loading_config")+"\n", dr.Index, dr.Name)
		for _, f := range dr.Fields {
			switch {
			case f.Skipped:
				fmt.Printf(locale.T("msg.field_skip")+"\n", f.Field)
			case f.Err != "":
				fmt.Printf(locale.T("msg.field_err")+"\n", f.Field, f.Err)
			default:
				fmt.Printf(locale.T("msg.field_ok")+"\n", f.Field)
			}
		}
	}
}

const (
	keyPowerLimit     = "power_limit"
	keyClockOffsetGPU = "clock_offset_gpu"
	keyClockOffsetMem = "clock_offset_mem"
	keyClockLimitGPU  = "clock_limit_gpu"
	keyFan            = "fan"
)

var orderedKeys = []string{
	keyPowerLimit, keyClockOffsetGPU, keyClockOffsetMem, keyClockLimitGPU, keyFan,
}

var paramNames = map[string][2]string{
	keyPowerLimit:     {"PowerLimit", "W"},
	keyClockOffsetGPU: {"ClockOffsetGPU", "MHz"},
	keyClockOffsetMem: {"ClockOffsetMem", "MHz"},
	keyClockLimitGPU:  {"ClockLimitGPU", "MHz"},
	keyFan:            {"Fan", "%"},
}

type daemonHandler struct {
	cfgCh      chan<- *config.Config
	backendErr error
}

func (h *daemonHandler) HandleLoad() ipc.LoadResponse {
	if h.backendErr != nil {
		return ipc.LoadResponse{Err: fmt.Sprintf("backend unavailable: %s", h.backendErr)}
	}
	cfg, err := config.Load()
	if err != nil {
		return ipc.LoadResponse{Err: err.Error()}
	}
	results := applyConfig(cfg)
	select {
	case h.cfgCh <- cfg:
	default:
	}
	return ipc.LoadResponse{Devices: results}
}

func (h *daemonHandler) HandleVersion() ipc.VersionResponse {
	if h.backendErr != nil {
		return ipc.VersionResponse{Err: fmt.Sprintf("backend unavailable: %s", h.backendErr)}
	}
	return ipc.VersionResponse{
		BackendName:    backend.Name(),
		BackendVersion: backend.Version(),
		DriverVersion:  backend.DriverVersion(),
	}
}

func (h *daemonHandler) HandleList(req ipc.ListRequest) ipc.ListResponse {
	if h.backendErr != nil {
		return ipc.ListResponse{Err: fmt.Sprintf("backend unavailable: %s", h.backendErr)}
	}
	targets, err := resolveIndices(req.Indices, req.All)
	if err != nil {
		return ipc.ListResponse{Err: err.Error()}
	}
	snaps := make([]gpu.Snapshot, 0, len(targets))
	for _, i := range targets {
		var s gpu.Snapshot
		s.Capture(devices[i])
		snaps = append(snaps, s)
	}
	return ipc.ListResponse{Devices: snaps}
}

func (h *daemonHandler) HandleTuneGet(req ipc.TuneGetRequest) ipc.TuneGetResponse {
	if h.backendErr != nil {
		return ipc.TuneGetResponse{Err: fmt.Sprintf("backend unavailable: %s", h.backendErr)}
	}
	targets, err := resolveIndices(req.Indices, req.All)
	if err != nil {
		return ipc.TuneGetResponse{Err: err.Error()}
	}
	var result []ipc.DeviceSpec
	for _, i := range targets {
		dev := devices[i]
		spec := ipc.DeviceSpec{Index: i, Name: dev.Name()}
		for _, k := range orderedKeys {
			spec.Rows = append(spec.Rows, buildSpecRow(dev, k))
		}
		result = append(result, spec)
	}
	return ipc.TuneGetResponse{Devices: result}
}

func (h *daemonHandler) HandleTuneSet(req ipc.TuneSetRequest) ipc.TuneSetResponse {
	if h.backendErr != nil {
		return ipc.TuneSetResponse{Err: fmt.Sprintf("backend unavailable: %s", h.backendErr)}
	}
	targets, err := resolveSetIndices(req.Indices, req.All)
	if err != nil {
		return ipc.TuneSetResponse{Err: err.Error()}
	}
	var result []ipc.DeviceResult
	for _, i := range targets {
		dev := devices[i]
		dr := ipc.DeviceResult{Index: i, Name: dev.Name()}
		for _, k := range orderedKeys {
			v, ok := req.Updates[k]
			if !ok {
				continue
			}
			p := paramNames[k]
			fr := ipc.FieldResult{Field: fmt.Sprintf("%s (%d%s)", p[0], v, p[1])}
			if err := applySet(dev, k, v); err != nil {
				fr.Err = err.Error()
			}
			dr.Fields = append(dr.Fields, fr)
		}
		result = append(result, dr)
	}
	return ipc.TuneSetResponse{Devices: result}
}

func (h *daemonHandler) HandleTuneReset(req ipc.TuneResetRequest) ipc.TuneResetResponse {
	if h.backendErr != nil {
		return ipc.TuneResetResponse{Err: fmt.Sprintf("backend unavailable: %s", h.backendErr)}
	}
	keys := req.Keys
	if len(keys) == 0 {
		keys = orderedKeys
	}
	targets, err := resolveIndices(req.Indices, req.All)
	if err != nil {
		return ipc.TuneResetResponse{Err: err.Error()}
	}
	var result []ipc.DeviceResult
	for _, i := range targets {
		dev := devices[i]
		dr := ipc.DeviceResult{Index: i, Name: dev.Name()}
		for _, k := range keys {
			fr := ipc.FieldResult{Field: fmt.Sprintf("reset %s", paramNames[k][0])}
			if err := applyReset(dev, k); err != nil {
				fr.Err = err.Error()
			}
			dr.Fields = append(dr.Fields, fr)
		}
		result = append(result, dr)
	}
	return ipc.TuneResetResponse{Devices: result}
}

func resolveIndices(indices []int, all bool) ([]int, error) {
	if all || len(indices) == 0 {
		out := make([]int, len(devices))
		for i := range devices {
			out[i] = i
		}
		return out, nil
	}
	for _, i := range indices {
		if i < 0 || i >= len(devices) {
			return nil, fmt.Errorf(locale.T("err.device_range"), i, len(devices)-1)
		}
	}
	return indices, nil
}

func resolveSetIndices(indices []int, all bool) ([]int, error) {
	if all {
		out := make([]int, len(devices))
		for i := range devices {
			out[i] = i
		}
		return out, nil
	}
	if len(indices) > 0 {
		for _, i := range indices {
			if i < 0 || i >= len(devices) {
				return nil, fmt.Errorf(locale.T("err.device_range"), i, len(devices)-1)
			}
		}
		return indices, nil
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("%s", locale.T("err.no_gpus"))
	}
	if len(devices) > 1 {
		return nil, fmt.Errorf("%s", locale.T("err.multi_gpu"))
	}
	return []int{0}, nil
}

const na = "N/A"

func fv(v int, err error) string {
	if err != nil || v == gpu.Unavailable {
		return na
	}
	return strconv.Itoa(v)
}

func fs(v int, err error) string {
	if err != nil || v == gpu.Unavailable {
		return na
	}
	return fmt.Sprintf("%+d", v)
}

func addUnit(s, unit string) string {
	if s == na || s == "auto" {
		return s
	}
	return s + unit
}

func buildSpecRow(dev gpu.Device, key string) ipc.SpecRow {
	row := ipc.SpecRow{Key: key}
	switch key {
	case keyPowerLimit:
		row.Name, row.Unit = "PowerLimit", "W"
		row.Current = fv(dev.PowerLimit())
		row.Default = fv(dev.PowerLimitDefault())
		mn, mx, err := dev.PowerLimitRange()
		row.Min, row.Max = fv(mn, err), fv(mx, err)
	case keyClockOffsetGPU:
		row.Name, row.Unit = "ClockOffsetGPU", "MHz"
		row.Current = fs(dev.ClockOffsetGPU())
		row.Default = "+0"
		mn, mx, err := dev.ClockOffsetGPURange()
		row.Min, row.Max = fs(mn, err), fs(mx, err)
	case keyClockOffsetMem:
		row.Name, row.Unit = "ClockOffsetMem", "MHz"
		row.Current = fs(dev.ClockOffsetMem())
		row.Default = "+0"
		mn, mx, err := dev.ClockOffsetMemRange()
		row.Min, row.Max = fs(mn, err), fs(mx, err)
	case keyClockLimitGPU:
		row.Name, row.Unit = "ClockLimitGPU", "MHz"
		row.Current = fv(dev.ClockLimitGPU())
		mn, mx, err := dev.ClockLimitGPURange()
		row.Min, row.Max = fv(mn, err), fv(mx, err)
		row.Default = row.Max
	case keyFan:
		row.Name, row.Unit = "Fan", "%"
		pct, _, err := dev.FanSpeed()
		row.Current = fv(pct, err)
		row.Default, row.Min, row.Max = "auto", "0", "100"
	}
	row.Current = addUnit(row.Current, row.Unit)
	row.Default = addUnit(row.Default, row.Unit)
	row.Min = addUnit(row.Min, row.Unit)
	row.Max = addUnit(row.Max, row.Unit)
	return row
}

func applySet(dev gpu.Device, key string, val int) error {
	switch key {
	case keyPowerLimit:
		return dev.SetPowerLimit(val)
	case keyClockOffsetGPU:
		return dev.SetClockOffsetGPU(val)
	case keyClockOffsetMem:
		return dev.SetClockOffsetMem(val)
	case keyClockLimitGPU:
		return dev.SetClockLimitGPU(val)
	case keyFan:
		return dev.SetFanSpeed(val)
	default:
		return fmt.Errorf(locale.T("err.unknown_param"), key)
	}
}

func applyReset(dev gpu.Device, key string) error {
	switch key {
	case keyPowerLimit:
		return dev.ResetPowerLimit()
	case keyClockOffsetGPU:
		return dev.ResetClockOffsetGPU()
	case keyClockOffsetMem:
		return dev.ResetClockOffsetMem()
	case keyClockLimitGPU:
		return dev.ResetClockLimitGPU()
	case keyFan:
		return dev.ResetFanSpeed()
	default:
		return fmt.Errorf(locale.T("err.unknown_param"), key)
	}
}

func runFanLoop(quit <-chan struct{}, cfgCh <-chan *config.Config, cfg *config.Config) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-quit:
			return
		case c := <-cfgCh:
			cfg = c
		case <-ticker.C:
			adjustFans(cfg)
		}
	}
}

func applyConfig(cfg *config.Config) []ipc.DeviceResult {
	if cfg == nil {
		return nil
	}
	var results []ipc.DeviceResult
	for i, dev := range devices {
		profile, ok := cfg.Settings[fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())]
		if !ok {
			continue
		}
		results = append(results, ipc.DeviceResult{
			Index:  i,
			Name:   dev.Name(),
			Fields: applyProfileToDevice(dev, profile),
		})
	}
	return results
}

func applyProfileToDevice(dev gpu.Device, profile config.Profile) []ipc.FieldResult {
	var results []ipc.FieldResult

	for _, f := range []struct {
		key string
		val *int
	}{
		{keyPowerLimit, profile.PowerLimit},
		{keyClockOffsetGPU, profile.ClockOffsetGPU},
		{keyClockOffsetMem, profile.ClockOffsetMem},
		{keyClockLimitGPU, profile.ClockLimitGPU},
	} {
		p := paramNames[f.key]
		if f.val == nil {
			results = append(results, ipc.FieldResult{Field: p[0], Skipped: true})
			continue
		}
		fr := ipc.FieldResult{Field: fmt.Sprintf("%s (%d%s)", p[0], *f.val, p[1])}
		if err := applySet(dev, f.key, *f.val); err != nil {
			fr.Err = err.Error()
		}
		results = append(results, fr)
	}

	switch {
	case profile.FanControl != nil && *profile.FanControl && len(profile.FanCurve) > 0:
		if err := config.ValidateFanCurve(profile.FanCurve); err != nil {
			results = append(results, ipc.FieldResult{
				Field: "FanCurve",
				Err:   fmt.Sprintf("invalid: %s — daemon will not manage fans", err),
			})
		} else {
				pts := make([]string, len(profile.FanCurve))
			for i, p := range profile.FanCurve {
				pts[i] = fmt.Sprintf("%d°C→%d%%", p.Temp, p.Fan)
			}
			results = append(results, ipc.FieldResult{
				Field: fmt.Sprintf("FanCurve: %s", strings.Join(pts, ", ")),
			})
		}
	case profile.FanControl != nil && !*profile.FanControl:
		fr := ipc.FieldResult{Field: "FanCurve reset"}
		if err := dev.ResetFanSpeed(); err != nil {
			fr.Err = err.Error()
		}
		results = append(results, fr)
	default:
		results = append(results, ipc.FieldResult{Field: "FanCurve", Skipped: true})
	}

	return results
}

func adjustFans(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for _, dev := range devices {
		profile, ok := cfg.Settings[fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())]
		if !ok || profile.FanControl == nil || !*profile.FanControl || len(profile.FanCurve) == 0 {
			continue
		}
		if config.ValidateFanCurve(profile.FanCurve) != nil {
			continue
		}
		temp, err := dev.Temperature()
		if err != nil {
			continue
		}
		_ = dev.SetFanSpeed(config.InterpolateFan(profile.FanCurve, temp))
	}
}

func resetFans(cfg *config.Config) {
	if cfg == nil {
		return
	}
	for _, dev := range devices {
		profile, ok := cfg.Settings[fmt.Sprintf("%s | %s", dev.Name(), dev.PCIBusID())]
		if ok && profile.FanControl != nil && *profile.FanControl {
			_ = dev.ResetFanSpeed()
		}
	}
}
