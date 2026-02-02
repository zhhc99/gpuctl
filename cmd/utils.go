package cmd

import (
	"fmt"
	"gpuctl/internal/gpu"
	"strconv"
	"strings"
)

// resolveDevices 按 -d flags 筛选设备.
// 无 flag 时返回 all.
func resolveDevices() ([]gpu.Device, error) {
	if len(deviceFlag) == 0 {
		return Devices, nil
	}

	var selected []gpu.Device
	for _, idStr := range deviceFlag {
		idx, err := parseDeviceID(idStr)
		if err != nil {
			return nil, err
		}

		if idx < 0 || idx >= len(Devices) {
			// TODO
			// 暂时不考虑前缀
			return nil, fmt.Errorf("device index %d out of range (0-%d)", idx, len(Devices)-1)
		}
		selected = append(selected, Devices[idx])
	}

	return selected, nil
}

// parseDeviceID 处理 0 或 n:0 这样的 id.
func parseDeviceID(id string) (int, error) {
	parts := strings.Split(id, ":")
	valStr := id
	if len(parts) == 2 {
		// TODO
		// 暂时不匹配前缀 parts[0], 现在只有 NVML 一个 backend.
		valStr = parts[1]
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("invalid device id format: %s", id)
	}
	return val, nil
}
