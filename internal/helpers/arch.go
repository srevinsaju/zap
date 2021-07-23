package helpers

import (
	"runtime"
	"strings"
)

var ARCH = map[string][]string{
	"amd64": {"x86_64", "amd64"},
	"386":   {"i386", "i686"},
	"arm":   {"arm"},
	"arm64": {"armhf", "aarch64"},
}

func HasArch(name string) bool {
	arch := ARCH[runtime.GOARCH]
	for i := range arch {
		if strings.Contains(name, arch[i]) {
			return true
		}
	}
	return false
}
