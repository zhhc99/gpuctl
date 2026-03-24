package cli

import (
	"fmt"

	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

var Version = "dev" // set via: -ldflags "-X 'github.com/zhhc99/gpuctl/internal/cli.Version=v1.0.0'"

func runVersion() error {
	fmt.Printf("gpuctl %s\n", Version)
	resp, err := ipc.PostVersion()
	if err != nil {
		fmt.Printf("%s", locale.T("msg.version_no_backend")+"\n")
		return nil
	}
	if resp.Err != "" {
		fmt.Printf(locale.T("msg.version_backend_err")+"\n", resp.Err)
		return nil
	}
	fmt.Printf(locale.T("msg.version_backend")+"\n",
		resp.BackendName, resp.BackendVersion, resp.DriverVersion)
	return nil
}
