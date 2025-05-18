> [!NOTE]
> This project is not associated with [nvm](https://github.com/nvm-sh/nvm) in any way. It is a personal project built for experimentation and learning.

## Installation

> [!IMPORTANT]
> The install script requires [`jq`](https://github.com/jqlang/jq) to be installed and available in your `$PATH`.

### Manual Installation

Navigate to the [Releases](https://github.com/aronhoyer/go-nvm/releases) page and download the artifact that matches your operating system and CPU architecture.

### Scripted Installation (Linux and macOS)

If you're on Linux or macOS, you can run the install script:

```sh
curl -s https://raw.githubusercontent.com/aronhoyer/go-nvm/refs/heads/main/install.sh | bash
```

To install the latest **pre-release** version, run:

```sh
curl -s https://raw.githubusercontent.com/aronhoyer/go-nvm/refs/heads/main/install.sh | bash -s -- --unstable
```
