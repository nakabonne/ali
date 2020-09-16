# ali

<div align="center">

Another load testing tool, inspired by [vegeta](https://github.com/tsenart/vegeta) and [jplot](https://github.com/rs/jplot).

![Screenshot](images/demo.gif)

</div>

`ali` comes with a simple terminal based UI, lets you generate HTTP requests and plot the results in real-time. With the help of [vegeta](https://github.com/tsenart/vegeta), it's available as a versatile load testing tool.

## Installation

Executables are available through the [releases page](https://github.com/nakabonne/ali/releases).

## Usage

## Features

#### Plot in real-time
Currently it only plots Latencies, but in the near future more metrics will be drawn do A as well.
![Screenshot](images/real-time.gif)

#### Visualize the attack progress
This will help you during long tests.
![Screenshot](images/progress.gif)

#### Mouse support
![Screenshot](images/mouse-support.gif)


## Built with
- [mum4k/termdash](https://github.com/mum4k/termdash/wiki/Termbox-API)
  - [nsf/termbox-go](https://github.com/nsf/termbox-go)
- [vegeta](https://github.com/tsenart/vegeta)


## LoadMap
- Plot more metrics in real-time ([#2](https://github.com/nakabonne/ali/issues/2))
- Support more options for HTTP requests ([#1](https://github.com/nakabonne/ali/issues/1))
- Better UI
