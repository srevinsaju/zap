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
echo [~] Uninstalling zap
rm $(which zap)

# Installation Complete
echo
echo [~] Uninstallation Complete....
