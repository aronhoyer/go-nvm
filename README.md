> [!NOTE]
> This project is not associated with [nvm](https://github.com/nvm-sh/nvm) in any way. Only a pet project of mine.

## Installation

> [!IMPORTANT]
> The install script requires you have [`jq`](https://github.com/jqlang/jq) installed and in your $PATH

Navigate to the releases and download the artifact corresponding to your operating system and CPU architecture.

If you're on Linux or macOS, you have the luxury of running the install script:

```sh
curl -s https://raw.githubusercontent.com/aronhoyer/go-nvm/refs/heads/main/install.sh | bash
```

If you want to install the latest pre-release version, you can run

```sh
curl -s https://raw.githubusercontent.com/aronhoyer/go-nvm/refs/heads/main/install.sh | bash -s -- --unstable
```
