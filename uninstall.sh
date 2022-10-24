#!/bin/bash
# A Bash Script to uninstall zap

# Error Handling
# Pipefall instead of error checks to retain original error
set -euo pipefail

chmod +x uninstall.sh
echo
echo "#########################"
echo
echo  '~   ZAP Uninstaller  ~'
echo
echo "#########################"
echo

# Config
REPO="srevinsaju/zap"

# Uninstallation

if [ "$EUID" -eq 0 ]; then
	echo [~] Script is running as root.
	echo
	echo [~] Uninstalling System-Wide
	sudo rm /usr/local/bin/zap
	# Done
	echo [~] Done....
else
	echo [~] Script is not running as root user
	echo
	echo '[~] Uninstalling Locally from ~/.local/bin/'
	rm ~/.local/bin/zap
	# Done
	echo [~] Done....
fi

# Installation Complete
echo
echo [~] Uninstallation Complete....
