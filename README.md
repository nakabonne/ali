# ali

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/nakabonne/ali)

<div align="center">

Another load testing tool, inspired by [vegeta](https://github.com/tsenart/vegeta) and [jplot](https://github.com/rs/jplot).

![Screenshot](images/demo.gif)

</div>

`ali` comes with a simple terminal based UI, lets you generate HTTP requests and plot the results in real-time.
With it, real-time analysis can be performed on the terminal.

## Installation

Executables are available through the [releases page](https://github.com/nakabonne/ali/releases).

**With Homebrew**

```bash
brew install nakabonne/ali/ali
```

**With Go**

```bash
go get github.com/nakabonne/ali
```

**With Docker**

```bash
docker run --rm -it nakabonne/ali ali
```

## Usage
### Quickstart
Give the target URL and press Enter, then the attack will be launched with default options.

### Options

#### Rate Limit
The request rate per time unit to issue against the targets.
Give 0 then it will send requests as fast as possible.
Default is `50`.

#### Duration
The amount of time to issue requests to the targets. Give `0s` for an infinite attack. Press `Ctrl-C` to stop.
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
Default is `10s`.

#### Timeout
The timeout for each request. `0s` means to disable timeouts.
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
Default is `30s`

#### Method
An HTTP request method for each request.

#### Header
A request header to be sent.

#### Body
The file whose content will be set as the http request body.

## Features

### Plot in real-time
Currently it only plots latencies, but in the near future more metrics will be drawn as well.

![Screenshot](images/real-time.gif)

### Visualize the attack progress
This will help you during long tests.

![Screenshot](images/progress.gif)

### Mouse support
With the help of [mum4k/termdash](https://github.com/mum4k/termdash), intuitive operation is supported.

![Screenshot](images/mouse-support.gif)


## Built with
- [mum4k/termdash](https://github.com/mum4k/termdash)
  - [nsf/termbox-go](https://github.com/nsf/termbox-go)
- [tsenart/vegeta](https://github.com/tsenart/vegeta)


## LoadMap
- Plot more metrics in real-time ([#2](https://github.com/nakabonne/ali/issues/2))
- Support more options for HTTP requests ([#1](https://github.com/nakabonne/ali/issues/1))
- Better UI
