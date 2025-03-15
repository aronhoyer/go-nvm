#!/usr/bin/bash

case "$(uname -s)" in
	Linux)
		nvm_os="linux"
		;;
	Darwin)
		nvm_os="darwin"
		;;
	*)
		echo "$(uname -s) not currently supported" >&2
		exit 1
		;;
esac

case "$(uname -m)" in
	x86_64)
		nvm_arch="x64"
		;;
	arm64)
		nvm_arch="arm64"
		;;
	*)
		echo "$(uname -m) not currently supported" >&2
		exit 1
		;;
esac

if ! command -v jq &> /dev/null; then
	echo "command jq not installed" >&2
	exit 1
fi

nvm_latest_tag="$(curl -s https://api.github.com/repos/aronhoyer/go-nvm/tags | jq -r '.[0].name')"

NVMDIR="$HOME/.go-nvm"

if [[ ! -d $NVMDIR ]]; then
	mkdir -p $NVMDIR
fi

curl -sSL -o "$NVMDIR/nvm" "https://github.com/aronhoyer/go-nvm/releases/download/$nvm_latest_tag/$nvm_latest_tag-gonvm-$nvm_os-$nvm_arch"
chmod +x "$NVMDIR/nvm"

if [[ ! -e $NVMDIR/env ]]; then
	curl -o "$NVMDIR/env" "https://raw.githubusercontent.com/aronhoyer/go-nvm/refs/tags/$nvm_latest_tag/env.sh"
	chmod +x "$NVMDIR/env"
fi

# TODO: check if the below line isn't actually already in profile
case "$SHELL" in
	*zsh)
		echo -e "\n. \"\$HOME/.go-nvm/env\"" >> $HOME/.zshrc
		;;
	*bash)
		echo -e "\n. \"\$HOME/.go-nvm/env\"" >> $HOME/.bashrc
		;;
	*)
		echo "$SHELL not currently supported" >&2
		exit 1
		;;
esac
