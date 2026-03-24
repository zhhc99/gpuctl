package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zhhc99/gpuctl/internal/daemon"
	"github.com/zhhc99/gpuctl/internal/locale"
)

func Run(args []string) error {
	if len(args) == 0 {
		fmt.Println(locale.T("help.root"))
		return nil
	}

	cmd, rest := args[0], args[1:]

	switch cmd {
	case "list":
		return runList(rest)
	case "tune":
		return runTune(rest)
	case "conf":
		return runConf(rest)
	case "load":
		return runLoad()
	case "daemon":
		return daemon.Run()
	case "health":
		return doHealth()
	case "version":
		return runVersion()
	case "help", "--help", "-h":
		if len(rest) > 0 {
			fmt.Println(helpFor(rest[0]))
		} else {
			fmt.Println(locale.T("help.root"))
		}
		return nil
	default:
		fmt.Fprintf(os.Stderr, locale.T("err.unknown_cmd"), cmd)
		return fmt.Errorf("exit 1")
	}
}

func helpFor(cmd string) string {
	key := "help." + cmd
	s := locale.T(key)
	if s == key {
		return fmt.Sprintf(locale.T("err.unknown_cmd"), cmd)
	}
	return s
}

func parseDeviceFlags(args []string) (rest []string, indices []int, all bool, err error) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-a", "--all":
			all = true
		case "-d", "--device":
			i++
			if i >= len(args) {
				return nil, nil, false, fmt.Errorf(locale.T("err.bad_flag_val"), "-d", "missing value")
			}
			indices, err = parseIntList(args[i])
			if err != nil {
				return nil, nil, false, fmt.Errorf(locale.T("err.bad_flag_val"), "-d", err)
			}
		default:
			if strings.HasPrefix(args[i], "-d=") || strings.HasPrefix(args[i], "--device=") {
				val := strings.SplitN(args[i], "=", 2)[1]
				indices, err = parseIntList(val)
				if err != nil {
					return nil, nil, false, fmt.Errorf(locale.T("err.bad_flag_val"), "-d", err)
				}
			} else {
				rest = append(rest, args[i])
			}
		}
	}
	return rest, indices, all, nil
}

func parseIntList(s string) ([]int, error) {
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, nil
}
