package cli

import (
	"fmt"

	"github.com/zhhc99/gpuctl/internal/ipc"
	"github.com/zhhc99/gpuctl/internal/locale"
)

func runLoad() error {
	resp, err := ipc.PostLoad()
	if err != nil {
		return err
	}
	if resp.Err != "" {
		return fmt.Errorf(locale.T("err.daemon"), resp.Err)
	}
	for _, dr := range resp.Devices {
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
	return nil
}
