package daemon

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// CheckIfRunningSystemd returns true if pid 1 is (a symlink to) systemd,
// otherwise false
// https://github.com/probonopd/go-appimage/blob/23ad67c727fb762867fe96db06d600a7cdaf297d/src/appimaged/prerequisites.go#L396
func CheckIfRunningSystemd() bool {

	prc := exec.Command("ps", "-p", "1", "-o", "comm=")
	out, err := prc.Output()
	if err != nil {
		logger.Debug(prc.String())
		logger.Debug(err)
		return false
	}
	if strings.TrimSpace(string(out)) == "systemd" {
		return true
	}
	return false
}



// CheckIfInvokedBySystemd returns true if this process has been invoked
// by systemd directly or indirectly, false in case it hasn't, the system is not
// using systemd, or we are not sure
// https://github.com/probonopd/go-appimage/blob/23ad67c727fb762867fe96db06d600a7cdaf297d/src/appimaged/prerequisites.go#L396
func CheckIfInvokedBySystemd() bool {

	if CheckIfRunningSystemd() == false {
		log.Println("This system is not running systemd")
		return false
	}

	if _, ok := os.LookupEnv("LAUNCHED_BY_SYSTEMD"); ok {
		log.Println("Launched by systemd: LAUNCHED_BY_SYSTEMD is present")
		return true
	}
	log.Println("Probably not launched by systemd (please file an issue if this is wrong)")
	return false
}