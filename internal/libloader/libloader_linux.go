//go:build linux

package libloader

import (
	"github.com/ebitengine/purego"
)

func Load(lib string) (uintptr, error) {
	h, err := purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	return h, err
}

func Bind(lib uintptr, ptr any, name string) {
	addr, err := purego.Dlsym(lib, name)
	if err == nil && addr != 0 {
		purego.RegisterFunc(ptr, addr)
	}
}
