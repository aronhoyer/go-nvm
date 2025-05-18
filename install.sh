#!/bin/bash

set -eu

if ! command -v jq >/dev/null 2>&1; then
	echo -e "\e[1;31Error:\e[0m jq must be in your \$PATH" >&2
	exit 69
fi

NVMDIR="${NVMDIR:-$HOME/.nvm}"
NVMBIN="$NVMDIR/bin"

NVM_GET_UNSTABLE=1
case "${1:-}" in
	--unstable) NVM_GET_UNSTABLE=0
	;;
esac

NVM_LATEST_RELEASE=""

if [ $NVM_GET_UNSTABLE ]; then
	echo "Downloading latest unstable..."
	NVM_LATEST_RELEASE="$(curl -s "https://api.github.com/repos/aronhoyer/go-nvm/releases" | jq -cr 'first')"
else
	echo "Downloading latest stable..."
	NVM_LATEST_RELEASE="$(curl -s "https://api.github.com/repos/aronhoyer/go-nvm/releases/latest")"
fi

NVM_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
NVM_ARCH="$(uname -m)"

case "$NVM_ARCH" in
	x86_64) NVM_ARCH="amd64"
		;;
	aarch64 | arm64) NVM_ARCH="arm64"
		;;
	*) echo "Unsupported architecture: $NVM_ARCH"; exit 64
		;;
esac

NVM_TAG_NAME="$(echo $NVM_LATEST_RELEASE | jq -r .tag_name)"
NVM_RELEASE_NAME="$(echo "$NVM_LATEST_RELEASE" | jq -r .name)"
NVM_ARTIFACT="$(echo "$NVM_LATEST_RELEASE" | jq -r '.assets[] | select(.name == "nvm-'"$NVM_OS"'-'"$NVM_ARCH"'.tar.gz")')"
NVM_ARTIFACT_URL="$(echo "$NVM_ARTIFACT" | jq -r .browser_download_url)"
NVM_ARTIFACT_NAME="$(echo "$NVM_ARTIFACT" | jq -r .name)"

echo "Downloading $NVM_ARTIFACT_NAME from $NVM_ARTIFACT_URL..."

NVM_DOWNLOAD_TARGET="$(mktemp -d)"
trap 'rm -rf $NVM_DOWNLOAD_TARGET' EXIT

pushd "$NVM_DOWNLOAD_TARGET"
curl -O -sLf "$NVM_ARTIFACT_URL"

if [ -d "$NVMDIR" ]; then
	rm -rf "$NVMDIR"
fi

mkdir "$NVM_RELEASE_NAME"
tar -C "$NVM_RELEASE_NAME" -xzf "$NVM_ARTIFACT_NAME"

mkdir -p "$NVMDIR"
cp -f "$NVM_RELEASE_NAME"/* "$NVMDIR"
popd

echo "nvm $NVM_TAG_NAME installed into $NVMDIR"
