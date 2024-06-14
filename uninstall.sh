#!/bin/bash
# A Bash Script to uninstall zap

# Error Handling
# Pipefall instead of error checks to retain original error
set -euo pipefail

echo
echo "#########################"
echo
echo  '~   ZAP Uninstaller  ~'
echo
echo "#########################"
echo

# Uninstallation
echo [~] Uninstalling zap

# Checks if zap is installed and get the path
ZAP_PATH=$(which zap || true)
if [[ -z "$ZAP_PATH" ]]; then
    echo "[!] zap is not installed or not found in PATH."
    exit 1
fi

# Remove the zap binary
rm "$ZAP_PATH"

# Uninstallation Complete
echo
echo "[~] Uninstallation Complete...."
