# ali
[![codecov.io Code Coverage](https://img.shields.io/codecov/c/github/nakabonne/ali.svg?maxAge=2592000)](https://codecov.io/github/nakabonne/ali?branch=master)
[![Release](https://img.shields.io/github/release/nakabonne/ali.svg?color=orange)](https://github.com/nakabonne/ali/releases/latest)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/nakabonne/ali)

A load testing tool aimed to perform real-time analysis, inspired by [vegeta](https://github.com/tsenart/vegeta) and [jplot](https://github.com/rs/jplot).

![Screenshot](images/demo.gif)

`ali` comes with a terminal-based UI where you can plot the metrics in real-time, so lets you perform real-time analysis on the terminal.

## Installation

Binaries are available through the [releases page](https://github.com/nakabonne/ali/releases).

**Via Homebrew**

```bash
brew install nakabonne/ali/ali
```

**Via APT**

```bash
wget https://github.com/nakabonne/ali/releases/download/v0.3.0/ali_0.3.0_linux_amd64.deb
apt install ./ali_0.3.0_linux_amd64.deb
```

**Via RPM**

```bash
curl -OL https://github.com/nakabonne/ali/releases/download/v0.3.0/ali_0.3.0_linux_amd64.rpm
rpm -i ./ali_0.3.0_linux_amd64.rpm
```

**Via AUR**

Thanks to [orhun](https://github.com/orhun), it's available as [ali](https://aur.archlinux.org/packages/ali) in the Arch User Repository.
```bash
yay -S ali
```

**Via Go**

```bash
go get github.com/nakabonne/ali
```

**Via Docker**

```bash
docker run --rm -it nakabonne/ali ali
```

## Usage
### Quickstart

```bash
ali http://host.xz
```
Replace `http://host.xz` with the target you want to issue the requests to.
Press Enter when the UI appears, then the attack will be launched with default options.

### Options

```bash
ali -h
Usage:
  ali [flags] <target URL>

Flags:
  -b, --body string         A request body to be sent.
  -B, --body-file string    The path to file whose content will be set as the http request body.
      --debug               Run in debug mode.
  -d, --duration duration   The amount of time to issue requests to the targets. Give 0s for an infinite attack. (default 10s)
  -H, --header strings      A request header to be sent. Can be used multiple times to send multiple headers.
  -m, --method string       An HTTP request method for each request. (default "GET")
  -r, --rate int            The request rate per second to issue against the targets. Give 0 then it will send requests as fast as possible. (default 50)
  -t, --timeout duration    The timeout for each request. 0s means to disable timeouts. (default 30s)
  -v, --version             Print the current version.
```

Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

## Features

### Plot in real-time
Currently it only plots latencies, but in the near future more metrics will be drawn as well.

![Screenshot](images/real-time.gif)

### Visualize the attack progress
This will help you during long tests.

![Screenshot](images/progress.gif)

### Mouse support
With the help of [mum4k/termdash](https://github.com/mum4k/termdash), it's intuitive to operate.

![Screenshot](images/mouse-support.gif)


## Built with
- [mum4k/termdash](https://github.com/mum4k/termdash)
  - [nsf/termbox-go](https://github.com/nsf/termbox-go)
- [tsenart/vegeta](https://github.com/tsenart/vegeta)


## Roadmap
- [x] Eliminate field-based configuration and only support configuration through cli flags
- [ ] Support more options for HTTP requests ([#1](https://github.com/nakabonne/ali/issues/1))
- [ ] Plot more metrics in real-time ([#2](https://github.com/nakabonne/ali/issues/2))
