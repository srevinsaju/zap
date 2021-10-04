#!/bin/bash
# curl https://github.com/srevinsaju/zap/raw/main/install.sh | sh


# A Bash Script to auto fetch latest release from srevinsaju/zap releases

echo
echo "####################"
echo
echo  '~   ZAP Installer  ~'
echo
echo "####################"
echo




# Config
REPO="srevinsaju/zap"


# function
function error_check {
	if [ $? -ne 0 ]; then
		echo [!] Error in downloading zap....
		echo [!] Exiting.
		exit 1
	fi
}


# Architecture
if [ -z $ARCH ]; then
	MACHINE_ARCH="$(uname -m)"
	if [ $MACHINE_ARCH = "amd64" ]; then
		ARCH="amd64"
	elif [ $MACHINE_ARCH = "i386" ]; then
		ARCH="386"
	elif [ $MACHINE_ARCH = "i686" ]; then
		ARCH="386" # both are 32bit, should be compatible
	elif [ $MACHINE_ARCH = "aarch64" ]; then
		ARCH="arm64"
	elif [ $MACHINE_ARCH = "arm64" ]; then
		ARCH="arm64"
	elif [ $MACHINE_ARCH = "arm" ]; then
		ARCH="arm"
	fi
fi

error_check

echo [~] Platform Arch: $ARCH
echo

# Requirements
CURL_PATH="$(which curl)"
GREP_PATH="$(which grep)"


echo [~] Requirements Check:

# required: curl
if [ -z $CURL_PATH ]; then
	echo
	echo [!] Command curl is required. Please install curl.
	exit 1
else
	echo -e [OK] curl
fi

# required: grep
if [ -z $GREP_PATH ]; then
	echo
	echo [!] Command grep was not found. Please install package containing grep tool.
	exit 1
else
	echo -e [OK] grep
fi


# required: jq
if [ -f "./jq" ]; then
	JQ_PATH="./jq"
	chmod 755 $JQ_PATH
else
	JQ_PATH="$(which jq)"
fi


if [ -z $JQ_PATH ]; then
	echo
	echo [!] Command jq was not found. Please install jq from your package manager....
	echo
	echo refer to https://github.com/stedolan/jq for instruction and downloads.
	echo
	exit 1
else
	echo -e [OK] jq
fi

if [ "$(which wget)" ]; then
	echo -e [OK] wget \(optional\)
fi


error_check
echo


# Get releases
RELEASE_API_URL="https://api.github.com/repos/$REPO/releases"

echo [~] Getting Latest zap release....

RELEASES="$($CURL_PATH -sN $RELEASE_API_URL)"
# RELEASES="$(cat r.json)"

error_check

# List releases

COMPATIBLE_RELEASE="$(echo "$RELEASES" | $JQ_PATH -rc .[].assets[].browser_download_url | grep -m 1 "$ARCH")"

if [ -z $COMPATIBLE_RELEASE ]; then
	echo [!] No compatible releases found....
	exit 1
fi

error_check
echo

echo [~] Found latest zap version....
echo

# Download release
echo [~] Downloading....
echo '[>] Download URL:' $COMPATIBLE_RELEASE
echo

TEMP_FILE="$(mktemp)"

if [ -z "$(which wget)" ]; then
	echo [~] Using Curl
	echo
	$CURL_PATH -L $COMPATIBLE_RELEASE -o $TEMP_FILE
else
	echo [~] wget is available, using wget.
	echo
	wget $COMPATIBLE_RELEASE -O $TEMP_FILE
fi

error_check
echo

# Installation

# Root and No Root
if [ "$(whoami)" == "root"]; then
	echo [~] Script is running as root.
	echo
	echo [~] Installing System-Wide
	sudo mv $TEMP_FILE /usr/local/bin/zap
	sudo chmod +x /usr/local/bin/zap
	# Done
	echo [~] Done....
else
	echo [~] Script is not running as root user
	echo
	echo '[~] Installing Locally to ~/.local/bin/'
	mkdir -p ~/.local/bin
	mv $TEMP_FILE "~/.local/bin/zap"
	chmod +x ~/.local/bin/zap
	# Add to $PATH
	echo '[~] Adding ~/.local/bin to PATH'
	PATH="$PATH;~/.local/bin/"
	if [ -f ~/.zshrc ]; then
		echo '[~] Adding .local/bin to ~/.zshrc'
		echo "PATH=\$PATH;~/.local/bin/" >> ~/.zshrc
	fi
	if [ -f ~/.bashrc ]; then
		echo '[~] Adding .local/bin to ~/.bashrc'
		echo "PATH=\$PATH;~/.local/bin" >> ~/.bashrc
	fi
	# Done
	echo [~] Done....
fi

# Check errors
error_check

# Installation Complete
echo
echo [~] Installation Complete....
