#!/usr/bin/bash

export NVMDIR="${NVMDIR:-$HOME/.go-nvm}"
export NVMBIN="$NVMDIR/bin"

case ":${PATH}:" in
    *:"$NVMDIR":*)
        ;;
    *:"$NVMBIN":*)
        ;;
    *)
        export PATH="$NVMDIR:$NVMBIN:$PATH"
        ;;
esac
