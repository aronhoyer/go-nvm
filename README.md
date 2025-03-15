> [!NOTE]
> This project is not associated with [nvm](https://github.com/nvm-sh/nvm) in any way. Only a pet project of mine.

## Installation

Just build it from source (requires Go).

```sh
git clone https://github.com/aronhoyer/go-nvm.git && cd go-nvm
go build -o $HOME/.go-nvm/nvm ./cmd/nvm
cp ./env.sh $HOME/.go-nvm/env
chmod +x $HOME/.go-nvm/env
```
