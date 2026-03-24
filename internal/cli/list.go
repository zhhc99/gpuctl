package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/zhhc99/gpuctl/internal/gpu"
	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

const na = "N/A"

func runList(args []string) error {
	_, indices, all, err := parseDeviceFlags(args)
	if err != nil {
		return err
	}
	resp, err := ipc.PostList(ipc.ListRequest{Indices: indices, All: all})
	if err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf(locale.T("err.daemon"), resp.Err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTEMP\tFAN\tPOWER\tUTIL\tCLOCK\tMEMORY")
	for _, s := range resp.Devices {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			s.Index, s.Name,
			snapTemp(s), snapFan(s), snapPower(s),
			snapUtil(s), snapClock(s), snapMemory(s),
		)
	}
	w.Flush()
	return nil
}

func snapTemp(s gpu.Snapshot) string {
	if s.Temperature != gpu.Unavailable {
		return fmt.Sprintf("%d°C", s.Temperature)
	}
	return na
}

func snapFan(s gpu.Snapshot) string {
	if s.FanPct != gpu.Unavailable {
		if s.FanRPM != gpu.Unavailable {
			return fmt.Sprintf("%d%%/%drpm", s.FanPct, s.FanRPM)
		}
		return fmt.Sprintf("%d%%", s.FanPct)
	}
	return na
}

func snapPower(s gpu.Snapshot) string {
	if s.Power != gpu.Unavailable {
		return fmt.Sprintf("%dW", s.Power)
	}
	return na
}

func snapUtil(s gpu.Snapshot) string {
	if s.UtilizationGPU != gpu.Unavailable {
		return fmt.Sprintf("G:%d%% M:%d%%", s.UtilizationGPU, s.UtilizationMem)
	}
	return na
}

func snapClock(s gpu.Snapshot) string {
	if s.ClockGpu != gpu.Unavailable {
		return fmt.Sprintf("G:%d M:%d", s.ClockGpu, s.ClockMem)
	}
	return na
}

func snapMemory(s gpu.Snapshot) string {
	if s.MemTotal != gpu.Unavailable {
		return fmt.Sprintf("%.1f/%.1fGB",
			float64(s.MemUsed)/(1<<30),
			float64(s.MemTotal)/(1<<30),
		)
	}
	return na
}
