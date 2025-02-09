# GLS - Go-based LS Package

## Overview

gls is ls but in Go. It provides a variety of options to customize the display of files and directories, including sorting, filtering (fast search throughout current dir), and previewing file contents, and extra features traditional ls does not provide.

## Installation

Install via Go, build and move to path:

```sh
go get github.com/rinimisini112/gls
```

Build and move to your path

```sh
go build -o gls github.com/rinimisini112/gls
sudo mv gls /usr/local/bin
```

### Or install executable for your OS

Download the latest release from the [build dir] renamed to gls and move to your path. MacOS and Linux (Windows big nono :D)

## Usage

```sh
gls [options] [directories]
```

### Options

- type -h or --help to see the help menu