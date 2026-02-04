![act-logo](https://raw.githubusercontent.com/wiki/nektos/act/img/logo-150.png)

# Overview [![push](https://github.com/nektos/act/workflows/push/badge.svg?branch=master&event=push)](https://github.com/nektos/act/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/nektos/act)](https://goreportcard.com/report/github.com/nektos/act) [![awesome-runners](https://img.shields.io/badge/listed%20on-awesome--runners-blue.svg)](https://github.com/jonico/awesome-runners)

> "Think globally, `act` locally"

Run your [GitHub Actions](https://developer.github.com/actions/) locally! Why would you want to do this? Two reasons:

- **Fast Feedback** - Rather than having to commit/push every time you want to test out the changes you are making to your `.github/workflows/` files (or for any changes to embedded GitHub actions), you can use `act` to run the actions locally. The [environment variables](https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables#default-environment-variables) and [filesystem](https://help.github.com/en/actions/reference/virtual-environments-for-github-hosted-runners#filesystems-on-github-hosted-runners) are all configured to match what GitHub provides.
- **Local Task Runner** - I love [make](<https://en.wikipedia.org/wiki/Make_(software)>). However, I also hate repeating myself. With `act`, you can use the GitHub Actions defined in your `.github/workflows/` to replace your `Makefile`!

## ✨ Multi-Runtime Support

Act supports both **Docker** and **Podman** with automatic detection:
- 🔒 **Podman preferred** for enhanced security (rootless, daemonless)
- 🐳 **Docker fallback** for maximum compatibility
- 🔄 **Zero breaking changes** - existing workflows continue to work

> [!TIP]
> **Now Manage and Run Act Directly From VS Code!**<br/>
> Check out the [GitHub Local Actions](https://sanjulaganepola.github.io/github-local-actions-docs/) Visual Studio Code extension which allows you to leverage the power of `act` to run and test workflows locally without leaving your editor.

# How Does It Work?

When you run `act` it reads in your GitHub Actions from `.github/workflows/` and determines the set of actions that need to be run. It uses the container runtime API (Docker or Podman) to either pull or build the necessary images, as defined in your workflow files and finally determines the execution path based on the dependencies that were defined. Once it has the execution path, it then uses the container runtime to run containers for each action based on the images prepared earlier. The [environment variables](https://help.github.com/en/actions/configuring-and-managing-workflows/using-environment-variables#default-environment-variables) and [filesystem](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners#file-systems) are all configured to match what GitHub provides.

Let's see it in action with a [sample repo](https://github.com/cplee/github-actions-demo)!

![Demo](https://raw.githubusercontent.com/wiki/nektos/act/quickstart/act-quickstart-2.gif)

# Container Runtime Support

Act supports multiple container runtimes with automatic detection and seamless switching:

## 🚀 Automatic Runtime Detection

Act automatically detects and uses the best available container runtime:

1. **🔒 Podman preferred** - Better security (rootless, no daemon required)
2. **🐳 Docker fallback** - Broader compatibility and mature tooling  
3. **⚙️ User override** - Full control via CLI flags or environment variables

## Quick Start

```bash
# Automatic detection (Podman preferred if available)
act

# Force specific runtime
act --container-runtime=podman
act --container-runtime=docker

# Custom socket path  
act --container-socket=/run/user/$(id -u)/podman/podman.sock
```

## Podman Installation

```bash
# macOS
brew install podman
podman machine init
podman machine start

# Ubuntu/Debian
sudo apt-get install podman

# Fedora/RHEL/CentOS  
sudo dnf install podman
```

> **macOS Note**: Act automatically detects Podman machine sockets on macOS, including SSH-based connections. No manual configuration needed!

## Configuration Options

| CLI Flag | Environment Variable | Description |
|----------|---------------------|-------------|
| `--container-runtime` | `ACT_CONTAINER_RUNTIME` | Set runtime: `auto`, `docker`, `podman` |
| `--container-socket` | `ACT_CONTAINER_SOCKET` | Custom socket path |

## Benefits of Podman

- **🔐 Enhanced Security**: Rootless containers by default
- **🚫 Daemonless**: No background daemon required
- **⚡ Better Performance**: Lower resource overhead  
- **🐧 Linux Native**: Better systemd integration

> **Zero Breaking Changes**: All existing Docker workflows continue to work unchanged!

For detailed information, see [Podman Support Documentation](docs/PODMAN_SUPPORT.md).

# Act User Guide

Please look at the [act user guide](https://nektosact.com) for more documentation.

# Support

Need help? Ask in [discussions](https://github.com/nektos/act/discussions)!

# Contributing

Want to contribute to act? Awesome! Check out the [contributing guidelines](CONTRIBUTING.md) to get involved.

## Manually building from source

- Install Go tools 1.20+ - (<https://golang.org/doc/install>)
- Clone this repo `git clone git@github.com:nektos/act.git`
- Run unit tests with `make test`
- Build and install: `make install`
