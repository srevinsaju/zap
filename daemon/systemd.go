package daemon

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/adrg/xdg"
)

// https://github.com/probonopd/go-appimage/blob/23ad67c727fb762867fe96db06d600a7cdaf297d/src/appimaged/prerequisites.go#L463
func installSystemdService() error {
	var err error
	var pathToServiceDir string

	self, err := os.Executable()
	if err != nil {
		return err
	}

	home, _ := os.UserHomeDir()
	// Note that https://www.freedesktop.org/software/systemd/man/systemd.unit.html
	// says $XDG_CONFIG_HOME/systemd/user or $HOME/.config/systemd/user
	// Units of packages that have been installed in the home directory
	// ($XDG_CONFIG_HOME is used if set, ~/.config otherwise)

	if os.Getenv("XDG_CONFIG_HOME") != "" {
		log.Println("Creating $XDG_CONFIG_HOME/systemd/user/zapd.service")
		err = os.MkdirAll(xdg.ConfigHome+"/systemd/user/", os.ModePerm)
		if err != nil {
			return err
		}
		pathToServiceDir = xdg.ConfigHome + "/systemd/user/"

	} else {
		log.Println("Creating ~/.config/systemd/user/zapd.service")
		err = os.MkdirAll(home+"/.config/systemd/user/", os.ModePerm)
		if err != nil {
			return err
		}
		pathToServiceDir = home + "/.config/systemd/user/"

	}

	logger.Debugf("Found the self path to be at %s", self)
	systemdService := []byte(`[Unit]
Description=Zap Updater daemon
After=syslog.target network.target
[Service]
Type=simple
ExecStart=` + self + ` daemon
LimitNOFILE=65536
RestartSec=3
Restart=always
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=zapd
Environment=LAUNCHED_BY_SYSTEMD=1
[Install]
WantedBy=default.target`)

	logger.Debugf("Writing zapd.service to %s", pathToServiceDir)
	err = os.WriteFile(path.Join(pathToServiceDir, "zapd.service"), systemdService, 0644)
	if err != nil {
		return err
	}

	prc := exec.Command("systemctl", "--user", "daemon-reload")
	_, err = prc.CombinedOutput()
	if err != nil {
		logger.Warn(prc.String())
		logger.Warn(err)
		return err
	}
	return nil
}

// SetupToRunThroughSystemd checks if this process has been launched through
// systemd on a systemd system and takes appropriate measures if it has not,
// either because systemd was not yet set up to launch it, or because
// another (newer?) version has been launched manually by the user outside
// of systemd
// https://github.com/probonopd/go-appimage/blob/23ad67c727fb762867fe96db06d600a7cdaf297d/src/appimaged/prerequisites.go#L322
func SetupToRunThroughSystemd() error {

	// When this process is being launched, then check whether we have been
	// launched by systemd. If the system is using systemd this process has
	// not been launched through it, then we probably want to exit here and let
	// systemd launch appimaged. We need to set up systemd to be able to do that
	// in case it is not already set up this way.

	if CheckIfRunningSystemd() == false {
		logger.Warn("This system is not running systemd")
		logger.Warn("This is not a problem; skipping checking the systemd service")
		return nil
	}

	if CheckIfInvokedBySystemd() == false {

		logger.Debugf("Manually launched, not by systemd. Check if enabled in systemd...")

		if _, err := os.Stat("/etc/systemd/user/zapd.service"); os.IsNotExist(err) {
			log.Println("/etc/systemd/user/zapd.service does not exist")
			err = installSystemdService()
			if err != nil {
				return err
			}
		}

		prc := exec.Command("systemctl", "--user", "status", "zapd")
		out, err := prc.CombinedOutput()
		if err != nil {
			logger.Debug(out)
			logger.Debug(err)
			// Note that if the service is stopped, we get an error exit code
			// with "exit status 3", hence this must not be fatal here
		}
		output := strings.TrimSpace(string(out))

		if strings.Contains(output, " enabled; ") {
			logger.Infof("Restarting via systemd...")
			prc := exec.Command("systemctl", "--user", "restart", "zapd")
			_, err := prc.CombinedOutput()
			if err != nil {
				logger.Debug(prc.String())
				logger.Debug(err)
			}
		} else {
			logger.Infof("Enabling systemd service...")
			prc := exec.Command("systemctl", "--user", "enable", "zapd")
			_, err := prc.CombinedOutput()
			if err != nil {
				logger.Debug(prc.String())
				logger.Debug(err)
			}
			logger.Infof("Starting systemd service...")
			prc = exec.Command("systemctl", "--user", "restart", "zapd")
			_, err = prc.CombinedOutput()
			if err != nil {
				logger.Debug(prc.String())
				logger.Debug(err)
			} else {
				logger.Infof("Exiting...")
				return nil
			}
		}

	}
	return nil

}
