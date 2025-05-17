export NVMDIR="${NVMDIR:-$HOME/.nvm}"
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
