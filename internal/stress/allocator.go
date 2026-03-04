package stress

import (
	"fmt"
	"strconv"
	"strings"
)

func parseVRAMBytes(s string) (uint64, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch {
	case strings.HasSuffix(s, "g"):
		v, err := strconv.ParseFloat(s[:len(s)-1], 64)
		if err == nil {
			return uint64(v * (1 << 30)), nil
		}
	case strings.HasSuffix(s, "m"):
		v, err := strconv.ParseFloat(s[:len(s)-1], 64)
		if err == nil {
			return uint64(v * (1 << 20)), nil
		}
	}
	return 0, fmt.Errorf("invalid vram %q: use e.g. 512m or 2g", s)
}
